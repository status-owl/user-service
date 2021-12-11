package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/rs/zerolog"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/status-owl/user-service/pkg/service"
	"github.com/status-owl/user-service/pkg/store"
	"github.com/status-owl/user-service/pkg/transport"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {

	var (
		metricsPort = flag.Int("metrics-port", 8081, "http port serving metrics")
		httpPort    = flag.Int("http-port", 8080, "http port")
		devLogging  = flag.Bool("dev-logging", false, "enables dev logging")
		logLevel    = flag.String("log-level", "info", "default log level")
		mongoDbUri  = flag.String("mongodb-uri", "", "mongodb connection uri")
		natsUrl     = flag.String("nats-url", nats.DefaultURL, "URL for connection to NATS")
		help        = flag.Bool("help", false, "print usage and exit")
	)
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	logger, err := setUpLogger(*devLogging, *logLevel)
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("failed to set up the logger")
		os.Exit(-1)
	}

	natsConn, err := nats.Connect(*natsUrl)
	if err != nil {
		logger.Error().
			Err(err).
			Str("url", *natsUrl).
			Msg("failed to establish connection to NATS server")

		os.Exit(1)
	}

	js, err := natsConn.JetStream()
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to create a JetStream")

		os.Exit(1)
	}

	logger.Info().
		Str("url", *natsUrl).
		Msg("configured nats")

	info, err := js.StreamInfo("USERS")
	if err != nil {
		logger.Fatal().
			Err(err).
			Str("url", *natsUrl).
			Msg("failed to get stream info from nats server or stream doesn't exist")
	}

	if info == nil {
		info, err = js.AddStream(&nats.StreamConfig{
			Name:     "USERS",
			Subjects: []string{"USERS.*"},
		})

		if err != nil {
			logger.Fatal().
				Err(err).
				Str("url", *natsUrl).
				Msg("failed to create USERS streams on nats server")

			os.Exit(1)
		}

		logger.Info().
			Str("name", info.Cluster.Name).
			Strs("subjects", info.Config.Subjects).
			Msg("created stream")
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

	svc := service.NewService(userStore, logger, js)
	httpHandler := transport.NewHTTPHandler(svc, logger)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *httpPort))
	if err != nil {
		logger.Fatal().
			Err(err).
			Int("port", *httpPort).
			Msgf("failed to listen on port %d", *httpPort)

		os.Exit(1)
	}

	// metrics listener
	http.Handle("/metrics", promhttp.Handler())
	metricsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", *metricsPort))
	if err != nil {
		logger.Fatal().
			Err(err).
			Int("port", *metricsPort).
			Msgf("failed to listen omn port %d", *metricsPort)

		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		logger.Info().
			Int("port", *metricsPort).
			Msgf("listening on metrics port %d...", *metricsPort)

		if err := http.Serve(metricsListener, http.DefaultServeMux); err != nil {
			logger.Fatal().
				Err(err).
				Int("port", *metricsPort).
				Msg("failed to start metrics http server")
		}
	}()

	go func() {
		defer wg.Done()
		logger.Info().
			Int("port", *httpPort).
			Msgf("listening on application port %d...", *httpPort)

		if err := http.Serve(listener, httpHandler); err != nil {
			logger.Fatal().
				Int("port", *httpPort).
				Err(err).
				Msg("failed to start application http server")
		}
	}()

	wg.Wait()
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

// setUpLogger create and configures a zerolog logger based on given parameters
func setUpLogger(dev bool, levelStr string) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		return zerolog.Nop(), fmt.Errorf("failed to determine log level '%s': %w", levelStr, err)
	}

	zerolog.SetGlobalLevel(level)
	var logger zerolog.Logger
	if dev {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			With().
			Caller().
			Timestamp().
			Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return logger, nil
}
