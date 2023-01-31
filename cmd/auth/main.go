package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-zoo/bone"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth/groups"
	gapi "github.com/mainflux/mainflux/auth/groups/api"
	gpostgres "github.com/mainflux/mainflux/auth/groups/postgres"
	gtracing "github.com/mainflux/mainflux/auth/groups/tracing"
	"github.com/mainflux/mainflux/auth/keys"
	kapi "github.com/mainflux/mainflux/auth/keys/api"
	"github.com/mainflux/mainflux/auth/keys/jwt"
	kpostgres "github.com/mainflux/mainflux/auth/keys/postgres"
	ktracing "github.com/mainflux/mainflux/auth/keys/tracing"
	"github.com/mainflux/mainflux/auth/policies"
	grpcapi "github.com/mainflux/mainflux/auth/policies/api/grpc"
	papi "github.com/mainflux/mainflux/auth/policies/api/http"
	ppostgres "github.com/mainflux/mainflux/auth/policies/postgres"
	ptracing "github.com/mainflux/mainflux/auth/policies/tracing"
	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	svcName       = "auth"
	stopWaitTime  = 5 * time.Second
	httpProtocol  = "http"
	httpsProtocol = "https"

	defLogLevel      = "debug"
	defSecretKey     = "clientsecret"
	defDBHost        = "localhost"
	defDBPort        = "5432"
	defDBUser        = "mainflux"
	defDBPass        = "mainflux"
	defDB            = "auth"
	defDBSSLMode     = "disable"
	defDBSSLCert     = ""
	defDBSSLKey      = ""
	defDBSSLRootCert = ""
	defHTTPPort      = "8180"
	defGRPCPort      = "8181"
	defSecret        = "auth"
	defServerCert    = ""
	defServerKey     = ""
	defJaegerURL     = "http://localhost:6831"
	defLoginDuration = "10h"

	envLogLevel      = "MF_AUTH_LOG_LEVEL"
	envSecretKey     = "MF_AUTH_SECRET_KEY"
	envDBHost        = "MF_AUTH_DB_HOST"
	envDBPort        = "MF_AUTH_DB_PORT"
	envDBUser        = "MF_AUTH_DB_USER"
	envDBPass        = "MF_AUTH_DB_PASS"
	envDB            = "MF_AUTH_DB"
	envDBSSLMode     = "MF_AUTH_DB_SSL_MODE"
	envDBSSLCert     = "MF_AUTH_DB_SSL_CERT"
	envDBSSLKey      = "MF_AUTH_DB_SSL_KEY"
	envDBSSLRootCert = "MF_AUTH_DB_SSL_ROOT_CERT"
	envHTTPPort      = "MF_AUTH_HTTP_PORT"
	envGRPCPort      = "MF_AUTH_GRPC_PORT"
	envSecret        = "MF_AUTH_SECRET"
	envServerCert    = "MF_AUTH_SERVER_CERT"
	envServerKey     = "MF_AUTH_SERVER_KEY"
	envJaegerURL     = "MF_JAEGER_URL"
	envLoginDuration = "MF_AUTH_LOGIN_TOKEN_DURATION"
)

type config struct {
	logLevel      string
	secretKey     string
	dbConfig      postgres.Config
	httpPort      string
	grpcPort      string
	serverCert    string
	serverKey     string
	jaegerURL     string
	loginDuration time.Duration
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	kdb, gdb, pdb := connectToDB(cfg.dbConfig, logger)
	defer kdb.Close()
	defer gdb.Close()
	defer pdb.Close()

	tp, err := initJaeger(svcName, cfg.jaegerURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger: %s", err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := otel.Tracer(svcName)

	ksvc, gsvc, psvc := newService(kdb, gdb, pdb, tracer, cfg, logger)

	g.Go(func() error {
		return startHTTPServer(ctx, ksvc, gsvc, psvc, cfg.httpPort, cfg.serverCert, cfg.serverKey, logger)
	})
	g.Go(func() error {
		return startGRPCServer(ctx, ksvc, psvc, cfg.grpcPort, cfg.serverCert, cfg.serverKey, logger)
	})

	g.Go(func() error {
		if sig := errors.SignalHandler(ctx); sig != nil {
			cancel()
			logger.Info(fmt.Sprintf("Authentication service shutdown by signal: %s", sig))
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Authentication service terminated: %s", err))
	}
}

func loadConfig() config {
	dbConfig := postgres.Config{
		Host:        mainflux.Env(envDBHost, defDBHost),
		Port:        mainflux.Env(envDBPort, defDBPort),
		User:        mainflux.Env(envDBUser, defDBUser),
		Pass:        mainflux.Env(envDBPass, defDBPass),
		Name:        mainflux.Env(envDB, defDB),
		SSLMode:     mainflux.Env(envDBSSLMode, defDBSSLMode),
		SSLCert:     mainflux.Env(envDBSSLCert, defDBSSLCert),
		SSLKey:      mainflux.Env(envDBSSLKey, defDBSSLKey),
		SSLRootCert: mainflux.Env(envDBSSLRootCert, defDBSSLRootCert),
	}

	loginDuration, err := time.ParseDuration(mainflux.Env(envLoginDuration, defLoginDuration))
	if err != nil {
		log.Fatal(err)
	}

	return config{
		logLevel:      mainflux.Env(envLogLevel, defLogLevel),
		secretKey:     mainflux.Env(envSecretKey, defSecretKey),
		dbConfig:      dbConfig,
		httpPort:      mainflux.Env(envHTTPPort, defHTTPPort),
		grpcPort:      mainflux.Env(envGRPCPort, defGRPCPort),
		serverCert:    mainflux.Env(envServerCert, defServerCert),
		serverKey:     mainflux.Env(envServerKey, defServerKey),
		jaegerURL:     mainflux.Env(envJaegerURL, defJaegerURL),
		loginDuration: loginDuration,
	}

}

func initJaeger(svcName, url string) (*tracesdk.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
		tracesdk.WithSpanProcessor(tracesdk.NewBatchSpanProcessor(exporter)),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(svcName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

func connectToDB(dbConfig postgres.Config, logger logger.Logger) (keys *sqlx.DB, groups *sqlx.DB, policies *sqlx.DB) {
	kdb, err := kpostgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	fmt.Println("KEYS")
	gdb, err := gpostgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	fmt.Println("GROUPS")
	pdb, err := ppostgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	fmt.Println("POLICIES")
	return kdb, gdb, pdb
}

func newService(kdb *sqlx.DB, gdb *sqlx.DB, pdb *sqlx.DB, tracer trace.Tracer, c config, logger logger.Logger) (keys.Service, groups.Service, policies.Service) {
	kdatabase := postgres.NewDatabase(pdb, tracer)
	krepo := kpostgres.New(kdatabase)

	gdatabase := postgres.NewDatabase(pdb, tracer)
	grepo := gpostgres.NewGroupRepo(gdatabase)

	pdatabase := postgres.NewDatabase(pdb, tracer)
	prepo := ppostgres.NewPolicyRepo(pdatabase)

	idp := uuid.New()
	tokenizer := jwt.New(c.secretKey)

	ksvc := keys.NewService(krepo, idp, tokenizer, c.loginDuration)
	gsvc := groups.NewService(krepo, grepo, prepo, tokenizer, idp)
	psvc := policies.NewService(prepo, tokenizer, krepo, idp)

	ksvc = ktracing.TracingMiddleware(ksvc, tracer)
	ksvc = kapi.LoggingMiddleware(ksvc, logger)
	ksvc = kapi.MetricsMiddleware(
		ksvc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "keys",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "keys",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	gsvc = gtracing.TracingMiddleware(gsvc, tracer)
	gsvc = gapi.LoggingMiddleware(gsvc, logger)
	gsvc = gapi.MetricsMiddleware(
		gsvc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "groups",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "groups",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	psvc = ptracing.TracingMiddleware(psvc, tracer)
	psvc = papi.LoggingMiddleware(psvc, logger)
	psvc = papi.MetricsMiddleware(
		psvc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "policies",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "policies",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	return ksvc, gsvc, psvc
}

func startHTTPServer(ctx context.Context, ksvc keys.Service, gsvc groups.Service, psvc policies.Service, port string, certFile string, keyFile string, logger logger.Logger) error {
	p := fmt.Sprintf(":%s", port)
	errCh := make(chan error)
	m := bone.New()
	kapi.MakeHandler(ksvc, m, logger)
	gapi.MakeGroupsHandler(gsvc, m, logger)
	papi.MakePolicyHandler(psvc, m, logger)
	server := &http.Server{Addr: p, Handler: m}

	protocol := httpProtocol
	switch {
	case certFile != "" || keyFile != "":
		logger.Info(fmt.Sprintf("Authentication service started using https, cert %s key %s, exposed port %s", certFile, keyFile, port))
		go func() {
			errCh <- server.ListenAndServeTLS(certFile, keyFile)
		}()
		protocol = httpsProtocol
	default:
		logger.Info(fmt.Sprintf("Authentication service started using http, exposed port %s", port))
		go func() {
			errCh <- server.ListenAndServe()
		}()
	}

	select {
	case <-ctx.Done():
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), stopWaitTime)
		defer cancelShutdown()
		if err := server.Shutdown(ctxShutdown); err != nil {
			logger.Error(fmt.Sprintf("Authentication %s service error occurred during shutdown at %s: %s", protocol, p, err))
			return fmt.Errorf("Authentication %s service error occurred during shutdown at %s: %w", protocol, p, err)
		}
		logger.Info(fmt.Sprintf("Authentication %s service shutdown of http at %s", protocol, p))
		return nil
	case err := <-errCh:
		return err
	}
}

func startGRPCServer(ctx context.Context, ksvc keys.Service, psvc policies.Service, port string, certFile string, keyFile string, logger logger.Logger) error {
	p := fmt.Sprintf(":%s", port)
	errCh := make(chan error)

	listener, err := net.Listen("tcp", p)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	var server *grpc.Server
	switch {
	case certFile != "" || keyFile != "":
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("failed to load auth certificates: %w", err)
		}
		logger.Info(fmt.Sprintf("Authentication gRPC service started using https on port %s with cert %s key %s", port, certFile, keyFile))
		server = grpc.NewServer(grpc.Creds(creds))
	default:
		logger.Info(fmt.Sprintf("Authentication gRPC service started using http on port %s", port))
		server = grpc.NewServer()
	}
	reflection.Register(server)
	mainflux.RegisterAuthServiceServer(server, grpcapi.NewServer(psvc, ksvc))
	logger.Info(fmt.Sprintf("Authentication gRPC service started, exposed port %s", port))
	go func() {
		errCh <- server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		c := make(chan bool)
		go func() {
			defer close(c)
			server.GracefulStop()
		}()
		select {
		case <-c:
		case <-time.After(stopWaitTime):
		}
		logger.Info(fmt.Sprintf("Authentication gRPC service shutdown at %s", p))
		return nil
	case err := <-errCh:
		return err
	}
}
