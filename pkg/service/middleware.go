package service

import (
	"context"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/log"
	"github.com/status-owl/user-service/pkg/model"
)

// Middleware describes a service middleware
type Middleware func(UserService) UserService

// LoggingMiddleware takes a logger as dependency
// and returns a service middleware
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next UserService) UserService {
		return &loggingMiddleware{log.WithPrefix(logger, "interface", "UserService"), next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   UserService
}

func (mw *loggingMiddleware) Create(ctx context.Context, user *RequestedUser) (id string, err error) {
	defer func() {
		mw.logger.Log(
			"method", "CreateUser",
			"user", user,
			"id", id,
			"err", err,
		)
	}()
	return mw.next.Create(ctx, user)
}

func (mw *loggingMiddleware) FindByID(ctx context.Context, id string) (user *model.User, err error) {
	defer func() {
		mw.logger.Log(
			"method", "FindUserByID",
			"id", id,
			"user", user,
			"err", err,
		)
	}()
	return mw.next.FindByID(ctx, id)
}

func InstrumentingMiddleware(createdUsers, fecthedUsers metrics.Counter) Middleware {
	return func(next UserService) UserService {
		return &instrumentigMiddleware{
			createdUsers: createdUsers,
			fetchedUsers: fecthedUsers,
			next:         next,
		}
	}
}

type instrumentigMiddleware struct {
	createdUsers metrics.Counter
	fetchedUsers metrics.Counter
	next         UserService
}

func (mw *instrumentigMiddleware) Create(ctx context.Context, user *RequestedUser) (string, error) {
	id, err := mw.next.Create(ctx, user)
	if err != nil {
		mw.createdUsers.With("status", "failed").Add(1)
	} else {
		mw.createdUsers.With("status", "failed").Add(1)
	}
	return id, err
}

func (mw *instrumentigMiddleware) FindByID(ctx context.Context, id string) (*model.User, error) {
	user, err := mw.next.FindByID(ctx, id)
	if err != nil {
		mw.fetchedUsers.With("status", "failed").Add(1)
	} else {
		mw.fetchedUsers.With("status", "success").Add(1)
	}
	return user, err
}
