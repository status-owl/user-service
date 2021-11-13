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

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/status-owl/user-service/pkg/endpoint"
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
		jsonLogger  = flag.Bool("json-logger", true, "if true - log as json")
		mongoDbUri  = flag.String("mongodb-uri", "", "mongodb connection uri")
		help        = flag.Bool("help", false, "print usage and exit")
	)
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	var logger log.Logger
	if *jsonLogger {
		logger = log.NewJSONLogger(os.Stdout)
	} else {
		logger = log.NewLogfmtLogger(os.Stdout)
	}

	mongoClient, err := connectMongo(*mongoDbUri)
	if err != nil {
		logger.Log("err", err, "msg", "failed to establish a connection to mongodb")
		os.Exit(1)
	}

	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			logger.Log("err", err, "msg", "failed to disconnect from mongodb")
			os.Exit(1)
		}
	}()

	userStore, err := store.NewUserStore(mongoClient, logger)
	if err != nil {
		logger.Log("err", err, "msg", "failed to create a user store")
		os.Exit(1)
	}

	tracer := opentracing.GlobalTracer()
	opentracing.SetGlobalTracer(tracer)

	var usersCreatedCounter, usersFetchedCounter metrics.Counter
	{
		usersCreatedCounter = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "status_owl",
			Subsystem: "user_service",
			Name:      "users_created",
			Help:      "Total count of created users",
		}, []string{"status"})

		usersFetchedCounter = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "status_owl",
			Subsystem: "user_service",
			Name:      "users_fetched",
			Help:      "Total count of fetched users",
		}, []string{"status"})
	}

	svc := service.NewService(userStore, logger, usersCreatedCounter, usersFetchedCounter)
	durationHistogram := prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "status_owl",
		Subsystem: "user_service",
		Name:      "request_duration_seconds",
		Help:      "Request duration in seconds.",
	}, []string{"method", "success"})

	endpoints := endpoint.NewEndpoints(svc, logger, durationHistogram, tracer)
	httpHandler := transport.NewHTTPHandler(endpoints, tracer, logger)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *httpPort))
	if err != nil {
		logger.Log("err", err, "msg", fmt.Sprintf("failed to listen on port %d", *httpPort))
		os.Exit(1)
	}

	// metrics listener
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())
	metricsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", *metricsPort))
	if err != nil {
		logger.Log("err", err, fmt.Sprintf("failed to listen omn port %d", *metricsPort))
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		logger.Log("msg", fmt.Sprintf("listening on metrics port %d...", *metricsPort))
		if err := http.Serve(metricsListener, http.DefaultServeMux); err != nil {
			logger.Log("err", err, "msg", "failed to start metrics http server")
		}
	}()

	go func() {
		defer wg.Done()
		logger.Log("msg", fmt.Sprintf("listening on application port %d...", *httpPort))
		if err := http.Serve(listener, httpHandler); err != nil {
			logger.Log("err", err, "msg", "failed to start application http server")
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
