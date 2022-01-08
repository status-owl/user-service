package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/konstantinwirz/srvgroup"
	"github.com/openzipkin/zipkin-go"
	"github.com/status-owl/user-service/pb"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/heptiolabs/healthcheck"

	"github.com/rs/zerolog"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/status-owl/user-service/pkg/service"
	"github.com/status-owl/user-service/pkg/store"
	"github.com/status-owl/user-service/pkg/transport"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	zipkinmiddleware "github.com/openzipkin/zipkin-go/middleware/http"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

var logger zerolog.Logger

func main() {
	var (
		httpPort    = flag.Int("http-port", 8080, "http port")
		metricsPort = flag.Int("metrics-port", 8081, "http port serving metrics")
		healthPort  = flag.Int("health-port", 8082, "port providing liveness and readiness endpoints")
		grpcPort    = flag.Int("grpc-port", 5000, "grpc server port")
		devLogging  = flag.Bool("dev-logging", false, "enables dev logging")
		logLevel    = flag.String("log-level", "info", "default log level")
		mongoDbUri  = flag.String("mongodb-uri", "", "mongodb connection uri")
		zipkinURL   = flag.String("zipkin-url", "", "Enable Zipkin tracing via HTTP reporter URL e.g. http://localhost:9411/api/v2/spans")
		help        = flag.Bool("help", false, "print usage and exit")
	)

	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if err := setUpLogger(*devLogging, *logLevel); err != nil {
		fmt.Printf("failed to initialize the logger: %s", err.Error())
		os.Exit(1)
	}

	mongoClient, err := connectMongo(*mongoDbUri)
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("failed to establish a connection to mongodb")

		os.Exit(1)
	}

	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			logger.Fatal().
				Err(err).
				Msg("failed to disconnect from mongodb")

			os.Exit(1)
		}
	}()

	userStore, err := store.NewUserStore(mongoClient, logger)
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("failed to create a user store")

		os.Exit(1)
	}

	tracer, cleanUpTracer, err := setUpTracer(*zipkinURL)
	defer cleanUpTracer()
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("failed to create zipkin tracer")
		os.Exit(1)
	}

	svc := service.NewService(userStore, logger)

	// set up application http server
	var appSrv srvgroup.Server
	{
		handler := transport.NewHTTPHandler(svc, logger)
		if tracer != nil {
			handler = zipkinmiddleware.NewServerMiddleware(tracer)(handler)
		}

		srv := http.Server{
			Addr:    fmt.Sprintf(":%d", *httpPort),
			Handler: handler,
		}

		appSrv = srvgroup.ServerLifecycleMiddleware(
			srvgroup.ServerLifecycleHooks{
				BeforeServe: func() {
					logger.Info().
						Str("address", srv.Addr).
						Msg("application http server listening...")
				}},
		)(srvgroup.HTTPServer(&srv))
	}

	// set up metrics http server
	var metricsSrv srvgroup.Server
	{
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		srv := http.Server{
			Addr:    fmt.Sprintf(":%d", *metricsPort),
			Handler: mux,
		}

		metricsSrv = srvgroup.ServerLifecycleMiddleware(
			srvgroup.ServerLifecycleHooks{
				BeforeServe: func() {
					logger.Info().
						Str("address", srv.Addr).
						Msg("metrics http server listening...")
				}},
		)(srvgroup.HTTPServer(&srv))
	}

	// set up grpc server
	var grpcSrv srvgroup.Server
	{
		var grpcServer *grpc.Server

		grpcSrv = srvgroup.Server{
			Serve: func() error {
				addr := fmt.Sprintf("localhost:%d", *grpcPort)
				lis, err := net.Listen("tcp", addr)
				if err != nil {
					return err
				}

				grpcServer = grpc.NewServer()
				pb.RegisterUserServiceServer(grpcServer, transport.NewBaseGrpcServer(svc))

				logger.Info().
					Str("address", addr).
					Msg("grpc application server listening...")

				if err = grpcServer.Serve(lis); err != nil {
					return err
				}
				return nil
			},
			Shutdown: func(ctx context.Context) error {
				if grpcServer != nil {
					grpcServer.GracefulStop()
				}
				return nil
			},
		}
	}

	// set up health http server
	var healthSrv srvgroup.Server
	{
		handler := healthcheck.NewHandler()
		handler.AddLivenessCheck(
			"http-server-check",
			healthcheck.TCPDialCheck(fmt.Sprintf(":%d", *httpPort), 1*time.Second),
		)
		handler.AddReadinessCheck(
			"mongodb-check",
			func() error { return pingMongo(mongoClient) },
		)

		srv := http.Server{
			Addr:    fmt.Sprintf(":%d", *healthPort),
			Handler: handler,
		}

		healthSrv = srvgroup.ServerLifecycleMiddleware(
			srvgroup.ServerLifecycleHooks{
				BeforeServe: func() {
					logger.Info().
						Str("address", srv.Addr).
						Msg("health http server listening...")
				}},
		)(srvgroup.HTTPServer(&srv))
	}

	for _, err := range srvgroup.Run(
		appSrv,
		metricsSrv,
		healthSrv,
		grpcSrv,
	) {
		logger.Error().
			Err(err).
			Send()
	}

	logger.Info().
		Msg("quit")
}

func pingMongo(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return client.Ping(ctx, readpref.Primary())
}

func connectMongo(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return mongoClient, pingMongo(mongoClient)
}

// setUpLogger creates and configures a zerolog logger based on given parameters
func setUpLogger(dev bool, levelStr string) error {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		logger = zerolog.Nop()
		return fmt.Errorf("failed to determine log level '%s': %w", levelStr, err)
	}

	zerolog.SetGlobalLevel(level)
	if dev {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			With().
			Caller().
			Timestamp().
			Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return nil
}

// setUpTracer creates and configures zipkin tracer
func setUpTracer(url string) (*zipkin.Tracer, func(), error) {
	var cleanUpFunc = func() {}
	reporter := zipkinhttp.NewReporter(url)
	cleanUpFunc = func() { _ = reporter.Close() }

	endpoint, err := zipkin.NewEndpoint("user-service", "localhost:0")
	if err != nil {
		return nil, cleanUpFunc, fmt.Errorf("failed to create zipkin's endpoint: %w", err)
	}

	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return nil, cleanUpFunc, fmt.Errorf("failed to create zipkin's tracer: %w", err)
	}

	return tracer, cleanUpFunc, nil
}
