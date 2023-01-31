// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-zoo/bone"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/policies/api/grpc"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/api"
	"github.com/mainflux/mainflux/users/bcrypt"
	"github.com/mainflux/mainflux/users/emailer"
	upostgres "github.com/mainflux/mainflux/users/postgres"
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
)

const (
	stopWaitTime = 5 * time.Second

	defLogLevel      = "debug"
	defDBHost        = "localhost"
	defDBPort        = "5432"
	defDBUser        = "mainflux"
	defDBPass        = "mainflux"
	defDB            = "users"
	defDBSSLMode     = "disable"
	defDBSSLCert     = ""
	defDBSSLKey      = ""
	defDBSSLRootCert = ""
	defHTTPPort      = "8180"
	defServerCert    = ""
	defServerKey     = ""
	defJaegerURL     = "http://localhost:6831"

	defEmailHost        = "localhost"
	defEmailPort        = "25"
	defEmailUsername    = "root"
	defEmailPassword    = ""
	defEmailFromAddress = ""
	defEmailFromName    = ""
	defEmailTemplate    = "email.tmpl"
	defAdminEmail       = "admin@example.com"
	defAdminPassword    = "12345678"
	defPassRegex        = "^.{8,}$"

	defTokenResetEndpoint = "/reset-request" // URL where user lands after click on the reset link from email

	defAuthTLS     = "false"
	defAuthCACerts = ""
	defAuthURL     = "localhost:8181"
	defAuthTimeout = "1s"

	defSelfRegister = "true" // By default, everybody can create a user. Otherwise, only admin can create a user.

	envLogLevel      = "MF_USERS_LOG_LEVEL"
	envDBHost        = "MF_USERS_DB_HOST"
	envDBPort        = "MF_USERS_DB_PORT"
	envDBUser        = "MF_USERS_DB_USER"
	envDBPass        = "MF_USERS_DB_PASS"
	envDB            = "MF_USERS_DB"
	envDBSSLMode     = "MF_USERS_DB_SSL_MODE"
	envDBSSLCert     = "MF_USERS_DB_SSL_CERT"
	envDBSSLKey      = "MF_USERS_DB_SSL_KEY"
	envDBSSLRootCert = "MF_USERS_DB_SSL_ROOT_CERT"
	envHTTPPort      = "MF_USERS_HTTP_PORT"
	envServerCert    = "MF_USERS_SERVER_CERT"
	envServerKey     = "MF_USERS_SERVER_KEY"
	envJaegerURL     = "MF_JAEGER_URL"

	envAdminEmail    = "MF_USERS_ADMIN_EMAIL"
	envAdminPassword = "MF_USERS_ADMIN_PASSWORD"
	envPassRegex     = "MF_USERS_PASS_REGEX"

	envEmailHost        = "MF_EMAIL_HOST"
	envEmailPort        = "MF_EMAIL_PORT"
	envEmailUsername    = "MF_EMAIL_USERNAME"
	envEmailPassword    = "MF_EMAIL_PASSWORD"
	envEmailFromAddress = "MF_EMAIL_FROM_ADDRESS"
	envEmailFromName    = "MF_EMAIL_FROM_NAME"
	envEmailTemplate    = "MF_EMAIL_TEMPLATE"

	envTokenResetEndpoint = "MF_TOKEN_RESET_ENDPOINT"

	envAuthTLS     = "MF_AUTH_CLIENT_TLS"
	envAuthCACerts = "MF_AUTH_CA_CERTS"
	envAuthURL     = "MF_AUTH_GRPC_URL"
	envAuthTimeout = "MF_AUTH_GRPC_TIMEOUT"

	envSelfRegister = "MF_USERS_ALLOW_SELF_REGISTER"
)

type config struct {
	logLevel      string
	dbConfig      postgres.Config
	emailConf     email.Config
	httpPort      string
	serverCert    string
	serverKey     string
	jaegerURL     string
	resetURL      string
	authTLS       bool
	authCACerts   string
	authURL       string
	authTimeout   time.Duration
	adminEmail    string
	adminPassword string
	passRegex     *regexp.Regexp
	selfRegister  bool
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}
	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	tp, err := initJaeger("users", cfg.jaegerURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger: %s", err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := otel.Tracer("users")

	auth, close := connectToAuth(cfg, logger)
	if close != nil {
		defer close()
	}

	svc := newService(db, tracer, auth, cfg, logger)

	g.Go(func() error {
		return startHTTPServer(ctx, svc, cfg.httpPort, cfg.serverCert, cfg.serverKey, logger)
	})

	g.Go(func() error {
		if sig := errors.SignalHandler(ctx); sig != nil {
			cancel()
			logger.Info(fmt.Sprintf("Users service shutdown by signal: %s", sig))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Users service terminated: %s", err))
	}
}

func loadConfig() config {
	authTimeout, err := time.ParseDuration(mainflux.Env(envAuthTimeout, defAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envAuthTimeout, err.Error())
	}

	tls, err := strconv.ParseBool(mainflux.Env(envAuthTLS, defAuthTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envAuthTLS)
	}

	passRegex, err := regexp.Compile(mainflux.Env(envPassRegex, defPassRegex))
	if err != nil {
		log.Fatalf("Invalid password validation rules %s\n", envPassRegex)
	}

	selfRegister, err := strconv.ParseBool(mainflux.Env(envSelfRegister, defSelfRegister))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envSelfRegister, err.Error())
	}

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

	emailConf := email.Config{
		FromAddress: mainflux.Env(envEmailFromAddress, defEmailFromAddress),
		FromName:    mainflux.Env(envEmailFromName, defEmailFromName),
		Host:        mainflux.Env(envEmailHost, defEmailHost),
		Port:        mainflux.Env(envEmailPort, defEmailPort),
		Username:    mainflux.Env(envEmailUsername, defEmailUsername),
		Password:    mainflux.Env(envEmailPassword, defEmailPassword),
		Template:    mainflux.Env(envEmailTemplate, defEmailTemplate),
	}

	return config{
		logLevel:      mainflux.Env(envLogLevel, defLogLevel),
		dbConfig:      dbConfig,
		emailConf:     emailConf,
		httpPort:      mainflux.Env(envHTTPPort, defHTTPPort),
		serverCert:    mainflux.Env(envServerCert, defServerCert),
		serverKey:     mainflux.Env(envServerKey, defServerKey),
		jaegerURL:     mainflux.Env(envJaegerURL, defJaegerURL),
		resetURL:      mainflux.Env(envTokenResetEndpoint, defTokenResetEndpoint),
		authTLS:       tls,
		authCACerts:   mainflux.Env(envAuthCACerts, defAuthCACerts),
		authURL:       mainflux.Env(envAuthURL, defAuthURL),
		authTimeout:   authTimeout,
		adminEmail:    mainflux.Env(envAdminEmail, defAdminEmail),
		adminPassword: mainflux.Env(envAdminPassword, defAdminPassword),
		passRegex:     passRegex,
		selfRegister:  selfRegister,
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

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := upostgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func connectToAuth(cfg config, logger logger.Logger) (mainflux.AuthServiceClient, func() error) {
	var opts []grpc.DialOption
	if cfg.authTLS {
		if cfg.authCACerts != "" {
			tpc, err := credentials.NewClientTLSFromFile(cfg.authCACerts, "")
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to create tls credentials: %s", err))
				os.Exit(1)
			}
			opts = append(opts, grpc.WithTransportCredentials(tpc))
		}
	} else {
		opts = append(opts, grpc.WithInsecure())
		logger.Info("gRPC communication is not encrypted")
	}

	conn, err := grpc.Dial(cfg.authURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to auth service: %s", err))
		os.Exit(1)
	}

	return authapi.NewClient(nil, conn, cfg.authTimeout), conn.Close
}

func newService(db *sqlx.DB, tracer trace.Tracer, auth mainflux.AuthServiceClient, c config, logger logger.Logger) users.Service {
	database := postgres.NewDatabase(db, tracer)
	urepo := upostgres.NewUsersRepo(database)

	hasher := bcrypt.New()

	emailer, err := emailer.New(c.resetURL, &c.emailConf)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to configure e-mailing util: %s", err.Error()))
	}

	idProvider := uuid.New()

	svc := users.New(urepo, hasher, auth, emailer, idProvider, c.passRegex)
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "users",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "users",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)
	if err := createAdmin(svc, urepo, hasher, c, auth); err != nil {
		logger.Error("failed to create admin user: " + err.Error())
		os.Exit(1)
	}

	return svc
}

func createAdmin(svc users.Service, urepo users.Repository, hsr users.Hasher, c config, auth mainflux.AuthServiceClient) error {
	id, err := uuid.New().ID()
	if err != nil {
		return err
	}
	hash, err := hsr.Hash(c.adminPassword)
	if err != nil {
		return err
	}

	user := users.User{
		ID:   id,
		Name: "admin",
		Credentials: users.Credentials{
			Identity: c.adminEmail,
			Secret:   hash,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    users.EnabledStatus,
	}

	if _, err := urepo.RetrieveByIdentity(context.Background(), user.Credentials.Identity); err == nil {
		return nil
	}

	// Create an admin
	if _, err = urepo.Save(context.Background(), user); err != nil {
		return err
	}
	user.Credentials.Secret = c.adminPassword
	_, err = svc.Login(context.Background(), user)
	if err != nil {
		return err
	}

	return nil
}

func startHTTPServer(ctx context.Context, svc users.Service, port string, certFile string, keyFile string, logger logger.Logger) error {
	p := fmt.Sprintf(":%s", port)
	errCh := make(chan error)
	m := bone.New()
	api.MakeClientsHandler(svc, m, logger)
	server := &http.Server{Addr: p, Handler: m}

	switch {
	case certFile != "" || keyFile != "":
		logger.Info(fmt.Sprintf("Users service started using https, cert %s key %s, exposed port %s", certFile, keyFile, port))
		go func() {
			errCh <- server.ListenAndServeTLS(certFile, keyFile)
		}()
	default:
		logger.Info(fmt.Sprintf("Users service started using http, exposed port %s", port))
		go func() {
			errCh <- server.ListenAndServe()
		}()
	}

	select {
	case <-ctx.Done():
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), stopWaitTime)
		defer cancelShutdown()
		if err := server.Shutdown(ctxShutdown); err != nil {
			logger.Error(fmt.Sprintf("Users service error occurred during shutdown at %s: %s", p, err))
			return fmt.Errorf("users service occurred during shutdown at %s: %w", p, err)
		}
		logger.Info(fmt.Sprintf("Users service shutdown of http at %s", p))
		return nil
	case err := <-errCh:
		return err
	}

}
