// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Package main contains coap-adapter main function to start the coap-adapter service.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"

	chclient "github.com/absmach/callhome/pkg/client"
	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/coap"
	"github.com/absmach/magistrala/coap/api"
	"github.com/absmach/magistrala/coap/tracing"
	"github.com/absmach/magistrala/internal"
	jaegerclient "github.com/absmach/magistrala/internal/clients/jaeger"
	"github.com/absmach/magistrala/internal/server"
	coapserver "github.com/absmach/magistrala/internal/server/coap"
	httpserver "github.com/absmach/magistrala/internal/server/http"
	mglog "github.com/absmach/magistrala/logger"
	"github.com/absmach/magistrala/pkg/auth"
	"github.com/absmach/magistrala/pkg/messaging/brokers"
	brokerstracing "github.com/absmach/magistrala/pkg/messaging/brokers/tracing"
	"github.com/absmach/magistrala/pkg/uuid"
	"github.com/caarlos0/env/v10"
	"github.com/grafana/loki-client-go/loki"
	"github.com/grafana/pyroscope-go"
	slogloki "github.com/samber/slog-loki/v3"
	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "coap_adapter"
	envPrefix      = "MG_COAP_ADAPTER_"
	envPrefixHTTP  = "MG_COAP_ADAPTER_HTTP_"
	envPrefixAuthz = "MG_THINGS_AUTH_GRPC_"
	defSvcHTTPPort = "5683"
	defSvcCoAPPort = "5683"
)

type config struct {
	LogLevel      string  `env:"MG_COAP_ADAPTER_LOG_LEVEL"   envDefault:"info"`
	BrokerURL     string  `env:"MG_MESSAGE_BROKER_URL"       envDefault:"nats://localhost:4222"`
	JaegerURL     url.URL `env:"MG_JAEGER_URL"               envDefault:"http://localhost:14268/api/traces"`
	SendTelemetry bool    `env:"MG_SEND_TELEMETRY"           envDefault:"true"`
	InstanceID    string  `env:"MG_COAP_ADAPTER_INSTANCE_ID" envDefault:""`
	TraceRatio    float64 `env:"MG_JAEGER_TRACE_RATIO"       envDefault:"1.0"`
	LokiURL       string  `env:"GOPHERCON_LOKI_URL"            envDefault:""`
	PyroScopeURL  string  `env:"GOPHERCON_PYROSCOPE_URL"       envDefault:""`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	var level slog.Level
	err := level.UnmarshalText([]byte(cfg.LogLevel))
	if err != nil {
		log.Fatalf("failed to parse log level: %s", err.Error())
	}
	fanout := slogmulti.Fanout(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		}),
	)
	if cfg.LokiURL != "" {
		config, err := loki.NewDefaultConfig(cfg.LokiURL)
		if err != nil {
			log.Fatalf("failed to create loki config: %s", err.Error())
		}
		config.TenantID = svcName
		client, err := loki.New(config)
		if err != nil {
			log.Fatalf("failed to create loki client: %s", err.Error())
		}

		hander := slogloki.Option{Level: level, Client: client}.NewLokiHandler()
		fanout = slogmulti.Fanout(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: level,
			}),
			hander,
		)
	}

	logger := slog.New(fanout).With("service", svcName)
	slog.SetDefault(logger)

	var exitCode int
	defer mglog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	coapServerConfig := server.Config{Port: defSvcCoAPPort}
	if err := env.ParseWithOptions(&coapServerConfig, env.Options{Prefix: envPrefix}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s CoAP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	authConfig := auth.Config{}
	if err := env.ParseWithOptions(&authConfig, env.Options{Prefix: envPrefixAuthz}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s auth configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	authClient, authHandler, err := auth.SetupAuthz(authConfig)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authHandler.Close()

	logger.Info("Successfully connected to things grpc server " + authHandler.Secure())

	tp, err := jaegerclient.NewProvider(ctx, svcName, cfg.JaegerURL, cfg.InstanceID, cfg.TraceRatio)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger: %s", err))
		exitCode = 1
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	if cfg.PyroScopeURL != "" {
		if _, err := pyroscope.Start(pyroscope.Config{
			ApplicationName: svcName,
			ServerAddress:   cfg.PyroScopeURL,
			Logger:          nil,
			ProfileTypes: []pyroscope.ProfileType{
				pyroscope.ProfileCPU,
				pyroscope.ProfileAllocObjects,
				pyroscope.ProfileAllocSpace,
				pyroscope.ProfileInuseObjects,
				pyroscope.ProfileInuseSpace,
				pyroscope.ProfileGoroutines,
				pyroscope.ProfileMutexCount,
			},
		}); err != nil {
			log.Fatalf("failed to start pyroscope: %s", err.Error())
		}
	}

	nps, err := brokers.NewPubSub(ctx, cfg.BrokerURL, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to message broker: %s", err))
		exitCode = 1
		return
	}
	defer nps.Close()
	nps = brokerstracing.NewPubSub(coapServerConfig, tracer, nps)

	svc := coap.New(authClient, nps)

	svc = tracing.New(tracer, svc)

	svc = api.LoggingMiddleware(svc, logger)

	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(cfg.InstanceID), logger)

	cs := coapserver.New(ctx, cancel, svcName, coapServerConfig, api.MakeCoAPHandler(svc, logger), logger)

	if cfg.SendTelemetry {
		chc := chclient.New(svcName, magistrala.Version, logger, cancel)
		go chc.CallHome(ctx)
	}

	g.Go(func() error {
		return hs.Start()
	})
	g.Go(func() error {
		return cs.Start()
	})
	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs, cs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("CoAP adapter service terminated: %s", err))
	}
}
