package service

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/status-owl/user-service/pkg/model"
)

// Middleware describes a service middleware
type Middleware func(UserService) UserService

// LoggingMiddleware takes a logger as dependency
// and returns a service middleware
func LoggingMiddleware(logger zerolog.Logger) Middleware {
	return func(next UserService) UserService {
		return &loggingMiddleware{
			logger.With().Str("interface", "UserService").Logger(),
			next,
		}
	}
}

type loggingMiddleware struct {
	logger zerolog.Logger
	next   UserService
}

func (mw *loggingMiddleware) Delete(ctx context.Context, id string) (err error) {
	logger := mw.logger.With().
		Str("method", "Delete").
		Str("id", id).
		Logger()

	logger.Trace().Msg("about to delete an user")

	defer func() {
		logger.Info().
			Err(err).
			Send()
	}()

	return mw.next.Delete(ctx, id)
}

func (mw *loggingMiddleware) Create(ctx context.Context, user *model.RequestedUser) (id string, err error) {
	logger := mw.logger.With().
		Str("method", "Create").
		Stringer("user", user).
		Logger()

	logger.Trace().Msg("about to create an user")

	defer func() {
		logger.Info().
			Str("id", id).
			Err(err).
			Send()
	}()

	return mw.next.Create(ctx, user)
}

func (mw *loggingMiddleware) FindByID(ctx context.Context, id string) (user *model.User, err error) {
	logger := mw.logger.With().
		Str("method", "FindByID").
		Str("id", id).
		Logger()

	logger.Trace().Msgf("about to find user %s", id)

	defer func() {
		logger.Info().
			Stringer("user", user).
			Err(err).
			Send()
	}()

	return mw.next.FindByID(ctx, id)
}

// Instrumenting Middleware

func InstrumentingMiddleware() Middleware {
	return func(next UserService) UserService {
		return &instrumentingMiddleware{
			createdUsers: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "status_owl",
				Subsystem: "user_service",
				Name:      "users_created",
				Help:      "Total count of created users",
			}, []string{"status"}),
			fetchedUsers: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "status_owl",
				Subsystem: "user_service",
				Name:      "users_fetched",
				Help:      "Total count of fetched users",
			}, []string{"status"}),
			deletedUsers: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "status_owl",
				Subsystem: "user_service",
				Name:      "users_deleted",
				Help:      "Total count of deleted users",
			}, []string{"status"}),
			next: next,
		}
	}
}

type instrumentingMiddleware struct {
	createdUsers, fetchedUsers, deletedUsers *prometheus.CounterVec
	next                                     UserService
}

func (mw *instrumentingMiddleware) Delete(ctx context.Context, id string) (err error) {
	defer func() {
		mw.deletedUsers.With(prometheus.Labels{"status": err2Status(err)}).Inc()
	}()

	err = mw.next.Delete(ctx, id)
	return
}

func err2Status(err error) string {
	if err != nil {
		return "failed"
	}
	return "success"
}

func (mw *instrumentingMiddleware) Create(ctx context.Context, user *model.RequestedUser) (id string, err error) {
	defer func() {
		mw.createdUsers.With(prometheus.Labels{"status": err2Status(err)}).Inc()
	}()

	id, err = mw.next.Create(ctx, user)
	return
}

func (mw *instrumentingMiddleware) FindByID(ctx context.Context, id string) (user *model.User, err error) {
	defer func() {
		mw.fetchedUsers.With(prometheus.Labels{"status": err2Status(err)}).Inc()
	}()

	user, err = mw.next.FindByID(ctx, id)
	return
}

func EventingMiddleware(js nats.JetStream) Middleware {
	return func(next UserService) UserService {
		return &eventingMiddleware{
			js:   js,
			next: next,
		}
	}
}

type eventingMiddleware struct {
	js   nats.JetStream
	once sync.Once
	next UserService
}

func (mw *eventingMiddleware) Delete(ctx context.Context, id string) error {
	err := mw.next.Delete(ctx, id)
	if err != nil {
		return err
	}

	var event = UserDeletedEvent{ID: id}
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = mw.js.Publish("USERS.deleted", b)
	if err != nil {
		return err
	}

	return nil
}

func (mw *eventingMiddleware) Create(ctx context.Context, user *model.RequestedUser) (string, error) {
	id, err := mw.next.Create(ctx, user)
	if err != nil {
		return id, err
	}

	var event = UserCreatedEvent{
		ID:  id,
		Pwd: string(user.Pwd),
	}

	b, err := json.Marshal(event)
	if err != nil {
		go func() {
			_ = mw.next.Delete(context.Background(), id)
		}()
		return "", err
	}

	_, err = mw.js.Publish("USERS.created", b)
	if err != nil {
		go func() {
			_ = mw.next.Delete(context.Background(), id)
		}()
		return "", err
	}

	return id, nil
}

func (mw *eventingMiddleware) FindByID(ctx context.Context, id string) (*model.User, error) {
	return mw.next.FindByID(ctx, id)
}

type UserCreatedEvent struct {
	ID  string `json:"id"`
	Pwd string `json:"pwd"`
}

type UserDeletedEvent struct {
	ID string `json:"id"`
}
