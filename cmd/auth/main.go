// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"time"

	chclient "github.com/absmach/callhome/pkg/client"
	"github.com/absmach/supermq"
	grpcAuthV1 "github.com/absmach/supermq/api/grpc/auth/v1"
	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	"github.com/absmach/supermq/auth"
	api "github.com/absmach/supermq/auth/api"
	authgrpcapi "github.com/absmach/supermq/auth/api/grpc/auth"
	tokengrpcapi "github.com/absmach/supermq/auth/api/grpc/token"
	httpapi "github.com/absmach/supermq/auth/api/http"
	"github.com/absmach/supermq/auth/bolt"
	"github.com/absmach/supermq/auth/hasher"
	"github.com/absmach/supermq/auth/jwt"
	apostgres "github.com/absmach/supermq/auth/postgres"
	"github.com/absmach/supermq/auth/tracing"
	boltclient "github.com/absmach/supermq/internal/clients/bolt"
	smqlog "github.com/absmach/supermq/logger"
	"github.com/absmach/supermq/pkg/jaeger"
	"github.com/absmach/supermq/pkg/policies/spicedb"
	pgclient "github.com/absmach/supermq/pkg/postgres"
	"github.com/absmach/supermq/pkg/prometheus"
	"github.com/absmach/supermq/pkg/server"
	grpcserver "github.com/absmach/supermq/pkg/server/grpc"
	httpserver "github.com/absmach/supermq/pkg/server/http"
	"github.com/absmach/supermq/pkg/uuid"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/caarlos0/env/v11"
	"github.com/jmoiron/sqlx"
	"go.etcd.io/bbolt"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	svcName        = "auth"
	envPrefixHTTP  = "SMQ_AUTH_HTTP_"
	envPrefixGrpc  = "SMQ_AUTH_GRPC_"
	envPrefixDB    = "SMQ_AUTH_DB_"
	envPrefixPATDB = "SMQ_AUTH_PAT_DB_"
	defDB          = "auth"
	defSvcHTTPPort = "8189"
	defSvcGRPCPort = "8181"
)

type config struct {
	LogLevel            string        `env:"SMQ_AUTH_LOG_LEVEL"               envDefault:"info"`
	SecretKey           string        `env:"SMQ_AUTH_SECRET_KEY"              envDefault:"secret"`
	JaegerURL           url.URL       `env:"SMQ_JAEGER_URL"                   envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry       bool          `env:"SMQ_SEND_TELEMETRY"               envDefault:"true"`
	InstanceID          string        `env:"SMQ_AUTH_ADAPTER_INSTANCE_ID"     envDefault:""`
	AccessDuration      time.Duration `env:"SMQ_AUTH_ACCESS_TOKEN_DURATION"   envDefault:"1h"`
	RefreshDuration     time.Duration `env:"SMQ_AUTH_REFRESH_TOKEN_DURATION"  envDefault:"24h"`
	InvitationDuration  time.Duration `env:"SMQ_AUTH_INVITATION_DURATION"     envDefault:"168h"`
	SpicedbHost         string        `env:"SMQ_SPICEDB_HOST"                 envDefault:"localhost"`
	SpicedbPort         string        `env:"SMQ_SPICEDB_PORT"                 envDefault:"50051"`
	SpicedbSchemaFile   string        `env:"SMQ_SPICEDB_SCHEMA_FILE"          envDefault:"./docker/spicedb/schema.zed"`
	SpicedbPreSharedKey string        `env:"SMQ_SPICEDB_PRE_SHARED_KEY"       envDefault:"12345678"`
	TraceRatio          float64       `env:"SMQ_JAEGER_TRACE_RATIO"           envDefault:"1.0"`
	ESURL               string        `env:"SMQ_ES_URL"                       envDefault:"amqp://guest:guest@rabbitmq:5672/"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err.Error())
	}

	logger, err := smqlog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err.Error())
	}

	var exitCode int
	defer smqlog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	dbConfig := pgclient.Config{Name: defDB}
	if err := env.ParseWithOptions(&dbConfig, env.Options{Prefix: envPrefixDB}); err != nil {
		logger.Error(err.Error())
	}

	am := apostgres.Migration()
	db, err := pgclient.Setup(dbConfig, *am)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer db.Close()

	tp, err := jaeger.NewProvider(ctx, svcName, cfg.JaegerURL, cfg.InstanceID, cfg.TraceRatio)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to init Jaeger: %s", err))
		exitCode = 1
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("error shutting down tracer provider: %v", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	spicedbclient, err := initSpiceDB(ctx, cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to init spicedb grpc client : %s\n", err.Error()))
		exitCode = 1
		return
	}

	boltDBConfig := boltclient.Config{}
	if err := env.ParseWithOptions(&boltDBConfig, env.Options{Prefix: envPrefixPATDB}); err != nil {
		logger.Error(fmt.Sprintf("failed to parse bolt db config : %s\n", err.Error()))
		exitCode = 1
		return
	}

	bClient, err := boltclient.Connect(boltDBConfig, bolt.Init)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to bolt db : %s\n", err.Error()))
		exitCode = 1
		return
	}
	defer bClient.Close()

	svc := newService(ctx, db, tracer, cfg, dbConfig, logger, spicedbclient, bClient, boltDBConfig)

	grpcServerConfig := server.Config{Port: defSvcGRPCPort}
	if err := env.ParseWithOptions(&grpcServerConfig, env.Options{Prefix: envPrefixGrpc}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s gRPC server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}
	registerAuthServiceServer := func(srv *grpc.Server) {
		reflection.Register(srv)
		grpcTokenV1.RegisterTokenServiceServer(srv, tokengrpcapi.NewTokenServer(svc))
		grpcAuthV1.RegisterAuthServiceServer(srv, authgrpcapi.NewAuthServer(svc))
	}

	gs := grpcserver.NewServer(ctx, cancel, svcName, grpcServerConfig, registerAuthServiceServer, logger)

	if cfg.SendTelemetry {
		chc := chclient.New(svcName, supermq.Version, logger, cancel)
		go chc.CallHome(ctx)
	}
	g.Go(func() error {
		return gs.Start()
	})

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}
	hs := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, httpapi.MakeHandler(svc, logger, cfg.InstanceID), logger)

	if cfg.SendTelemetry {
		chc := chclient.New(svcName, supermq.Version, logger, cancel)
		go chc.CallHome(ctx)
	}

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs, gs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("users service terminated: %s", err))
	}
}

func initSpiceDB(ctx context.Context, cfg config) (*authzed.ClientWithExperimental, error) {
	client, err := authzed.NewClientWithExperimentalAPIs(
		fmt.Sprintf("%s:%s", cfg.SpicedbHost, cfg.SpicedbPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(cfg.SpicedbPreSharedKey),
	)
	if err != nil {
		return client, err
	}

	if err := initSchema(ctx, client, cfg.SpicedbSchemaFile); err != nil {
		return client, err
	}

	return client, nil
}

func initSchema(ctx context.Context, client *authzed.ClientWithExperimental, schemaFilePath string) error {
	schemaContent, err := os.ReadFile(schemaFilePath)
	if err != nil {
		return fmt.Errorf("failed to read spice db schema file : %w", err)
	}

	if _, err = client.SchemaServiceClient.WriteSchema(ctx, &v1.WriteSchemaRequest{Schema: string(schemaContent)}); err != nil {
		return fmt.Errorf("failed to create schema in spicedb : %w", err)
	}

	return nil
}

func newService(_ context.Context, db *sqlx.DB, tracer trace.Tracer, cfg config, dbConfig pgclient.Config, logger *slog.Logger, spicedbClient *authzed.ClientWithExperimental, bClient *bbolt.DB, bConfig boltclient.Config) auth.Service {
	database := pgclient.NewDatabase(db, dbConfig, tracer)
	keysRepo := apostgres.New(database)
	patsRepo := bolt.NewPATSRepository(bClient, bConfig.Bucket)
	hasher := hasher.New()
	idProvider := uuid.New()

	pEvaluator := spicedb.NewPolicyEvaluator(spicedbClient, logger)
	pService := spicedb.NewPolicyService(spicedbClient, logger)

	t := jwt.New([]byte(cfg.SecretKey))

	svc := auth.New(keysRepo, patsRepo, hasher, idProvider, t, pEvaluator, pService, cfg.AccessDuration, cfg.RefreshDuration, cfg.InvitationDuration)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := prometheus.MakeMetrics("auth", "api")
	svc = api.MetricsMiddleware(svc, counter, latency)
	svc = tracing.New(svc, tracer)

	return svc
}
