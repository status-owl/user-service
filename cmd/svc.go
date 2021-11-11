package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
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

	logger := log.NewJSONLogger(os.Stdout)

	mongoClient, err := connectMongo("mongodb://root:secret@localhost:27017")
	if err != nil {
		logger.Log("err", err, "msg", "failed to establish a connection to mongodb")
		os.Exit(1)
	}

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
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		logger.Log("err", err, "msg", "failed to listen on port 8080")
		os.Exit(1)
	}

	httpHandler.(*mux.Router).Handle("/metrics", promhttp.Handler())
	logger.Log("msg", "listening on port 8080...")
	if err := http.Serve(listener, httpHandler); err != nil {
		logger.Log("err", err, "msg", "failed to start http server")
		os.Exit(1)
	}
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
