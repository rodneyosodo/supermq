// Copyright (c) Magistrala
// SPDX-License-Identifier: Apache-2.0

// Package main contains things main function to start the things service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/internal"
	authclient "github.com/absmach/magistrala/internal/clients/grpc/auth"
	jaegerclient "github.com/absmach/magistrala/internal/clients/jaeger"
	pgclient "github.com/absmach/magistrala/internal/clients/postgres"
	redisclient "github.com/absmach/magistrala/internal/clients/redis"
	"github.com/absmach/magistrala/internal/env"
	mfgroups "github.com/absmach/magistrala/internal/groups"
	gapi "github.com/absmach/magistrala/internal/groups/api"
	gpostgres "github.com/absmach/magistrala/internal/groups/postgres"
	gtracing "github.com/absmach/magistrala/internal/groups/tracing"
	"github.com/absmach/magistrala/internal/postgres"
	"github.com/absmach/magistrala/internal/server"
	grpcserver "github.com/absmach/magistrala/internal/server/grpc"
	httpserver "github.com/absmach/magistrala/internal/server/http"
	mflog "github.com/absmach/magistrala/logger"
	"github.com/absmach/magistrala/pkg/groups"
	"github.com/absmach/magistrala/pkg/uuid"
	"github.com/absmach/magistrala/things"
	"github.com/absmach/magistrala/things/api"
	grpcapi "github.com/absmach/magistrala/things/api/grpc"
	httpapi "github.com/absmach/magistrala/things/api/http"
	thcache "github.com/absmach/magistrala/things/cache"
	thevents "github.com/absmach/magistrala/things/events"
	thingspg "github.com/absmach/magistrala/things/postgres"
	localusers "github.com/absmach/magistrala/things/standalone"
	ctracing "github.com/absmach/magistrala/things/tracing"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	callhome "github.com/mainflux/callhome/pkg/client"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	svcName            = "things"
	envPrefixDB        = "MG_THINGS_DB_"
	envPrefixHTTP      = "MG_THINGS_HTTP_"
	envPrefixGRPC      = "MG_THINGS_AUTH_GRPC_"
	defDB              = "things"
	defSvcHTTPPort     = "9000"
	defSvcAuthGRPCPort = "7000"
)

type config struct {
	LogLevel         string  `env:"MG_THINGS_LOG_LEVEL"           envDefault:"info"`
	StandaloneID     string  `env:"MG_THINGS_STANDALONE_ID"       envDefault:""`
	StandaloneToken  string  `env:"MG_THINGS_STANDALONE_TOKEN"    envDefault:""`
	JaegerURL        string  `env:"MG_JAEGER_URL"                 envDefault:"http://jaeger:14268/api/traces"`
	CacheKeyDuration string  `env:"MG_THINGS_CACHE_KEY_DURATION"  envDefault:"10m"`
	SendTelemetry    bool    `env:"MG_SEND_TELEMETRY"             envDefault:"true"`
	InstanceID       string  `env:"MG_THINGS_INSTANCE_ID"         envDefault:""`
	ESURL            string  `env:"MG_THINGS_ES_URL"              envDefault:"redis://localhost:6379/0"`
	CacheURL         string  `env:"MG_THINGS_CACHE_URL"           envDefault:"redis://localhost:6379/0"`
	TraceRatio       float64 `env:"MG_JAEGER_TRACE_RATIO"         envDefault:"1.0"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// Create new things configuration
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := mflog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}

	var exitCode int
	defer mflog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	// Create new database for things
	dbConfig := pgclient.Config{Name: defDB}
	if err := dbConfig.LoadEnv(envPrefixDB); err != nil {
		logger.Fatal(err.Error())
	}

	tm := thingspg.Migration()
	gm := gpostgres.Migration()
	tm.Migrations = append(tm.Migrations, gm.Migrations...)
	db, err := pgclient.SetupWithConfig(envPrefixDB, *tm, dbConfig)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer db.Close()

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

	// Setup new redis cache client
	cacheclient, err := redisclient.Connect(cfg.CacheURL)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer cacheclient.Close()

	var auth magistrala.AuthServiceClient

	switch cfg.StandaloneID != "" && cfg.StandaloneToken != "" {
	case true:
		auth = localusers.NewAuthService(cfg.StandaloneID, cfg.StandaloneToken)
		logger.Info("Using standalone auth service")
	default:
		authServiceClient, authHandler, err := authclient.Setup(svcName)
		if err != nil {
			logger.Error(err.Error())
			exitCode = 1
			return
		}
		defer authHandler.Close()
		auth = authServiceClient
		logger.Info("Successfully connected to auth grpc server " + authHandler.Secure())
	}

	csvc, gsvc, err := newService(ctx, db, dbConfig, auth, cacheclient, cfg.CacheKeyDuration, cfg.ESURL, tracer, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create services: %s", err))
		exitCode = 1
		return
	}

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}
	mux := chi.NewRouter()
	httpSvc := httpserver.New(ctx, cancel, svcName, httpServerConfig, httpapi.MakeHandler(csvc, gsvc, mux, logger, cfg.InstanceID), logger)

	grpcServerConfig := server.Config{Port: defSvcAuthGRPCPort}
	if err := env.Parse(&grpcServerConfig, env.Options{Prefix: envPrefixGRPC}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s gRPC server configuration : %s", svcName, err))
		exitCode = 1
		return
	}
	regiterAuthzServer := func(srv *grpc.Server) {
		reflection.Register(srv)
		magistrala.RegisterAuthzServiceServer(srv, grpcapi.NewServer(csvc))
	}
	gs := grpcserver.New(ctx, cancel, svcName, grpcServerConfig, regiterAuthzServer, logger)

	if cfg.SendTelemetry {
		chc := callhome.New(svcName, magistrala.Version, logger, cancel)
		go chc.CallHome(ctx)
	}

	// Start all servers
	g.Go(func() error {
		return httpSvc.Start()
	})

	g.Go(func() error {
		return gs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, httpSvc)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
	}
}

func newService(ctx context.Context, db *sqlx.DB, dbConfig pgclient.Config, auth magistrala.AuthServiceClient, cacheClient *redis.Client, keyDuration, esURL string, tracer trace.Tracer, logger mflog.Logger) (things.Service, groups.Service, error) {
	database := postgres.NewDatabase(db, dbConfig, tracer)
	cRepo := thingspg.NewRepository(database)
	gRepo := gpostgres.New(database)

	idp := uuid.New()

	kDuration, err := time.ParseDuration(keyDuration)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to parse cache key duration: %s", err.Error()))
	}

	thingCache := thcache.NewCache(cacheClient, kDuration)

	csvc := things.NewService(auth, cRepo, gRepo, thingCache, idp)
	gsvc := mfgroups.NewService(gRepo, idp, auth)

	csvc, err = thevents.NewEventStoreMiddleware(ctx, csvc, esURL)
	if err != nil {
		return nil, nil, err
	}

	csvc = ctracing.New(csvc, tracer)
	csvc = api.LoggingMiddleware(csvc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	csvc = api.MetricsMiddleware(csvc, counter, latency)

	gsvc = gtracing.New(gsvc, tracer)
	gsvc = gapi.LoggingMiddleware(gsvc, logger)
	counter, latency = internal.MakeMetrics(fmt.Sprintf("%s_groups", svcName), "api")
	gsvc = gapi.MetricsMiddleware(gsvc, counter, latency)

	return csvc, gsvc, err
}
