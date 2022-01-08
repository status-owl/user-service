package service

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
			logger.With().
				Str("interface", "UserService").
				Logger(),
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
		if err != nil {
			logger.Error().
				Err(err).
				Msg("failed to delete an user")
		} else {
			logger.Info().
				Msg("deleted user")
		}
	}()

	return mw.next.Delete(ctx, id)
}

func (mw *loggingMiddleware) Create(ctx context.Context, user model.RequestedUser) (id string, err error) {
	logger := mw.logger.With().
		Str("method", "Create").
		Stringer("user", user).
		Logger()

	logger.Trace().Msg("about to create an user")

	defer func() {
		if err != nil {
			logger.Info().
				Err(err).
				Msg("failed to create an user")
		} else {
			logger.Info().
				Str("id", id).
				Msg("user created")
		}
	}()

	return mw.next.Create(ctx, user)
}

func (mw *loggingMiddleware) FindByID(ctx context.Context, id string) (user *model.User, err error) {
	logger := mw.logger.With().
		Str("method", "FindByID").
		Str("id", id).
		Logger()

	logger.Trace().Msg("about to find an user")

	defer func() {
		if err != nil {
			logger.Error().
				Err(err).
				Msg("failed to find an user")
		} else {
			logger.Info().
				Stringer("user", user).
				Err(err).
				Msg("user found")
		}
	}()

	user, err = mw.next.FindByID(ctx, id)
	return
}

// Instrumenting Middleware

func InstrumentingMiddleware() Middleware {
	return func(next UserService) UserService {
		return &instrumentingMiddleware{
			createdUsers: promauto.NewCounterVec(prometheus.CounterOpts{
				Namespace: "status_owl",
				Subsystem: "user_service",
				Name:      "users_created",
				Help:      "Total count of created users",
			}, []string{"status"}),
			fetchedUsers: promauto.NewCounterVec(prometheus.CounterOpts{
				Namespace: "status_owl",
				Subsystem: "user_service",
				Name:      "users_fetched",
				Help:      "Total count of fetched users",
			}, []string{"status"}),
			deletedUsers: promauto.NewCounterVec(prometheus.CounterOpts{
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

func (mw *instrumentingMiddleware) Create(ctx context.Context, user model.RequestedUser) (id string, err error) {
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
