package store

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/status-owl/user-service/pkg/model"
)

// contains logging middleware for the UserStore

type Middleware func(UserStore) UserStore

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next UserStore) UserStore {
		return &loggingMiddleware{
			logger: log.WithPrefix(logger, "interface", "UserStore"),
			next:   next,
		}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   UserStore
}

func (mw *loggingMiddleware) Create(ctx context.Context, user *model.User) (id string, err error) {
	l := log.WithPrefix(mw.logger, "method", "Create")
	level.Debug(l).Log("user", user, "msg", "about to create a user")

	defer func(begin time.Time) {
		level.Info(l).Log(
			"user", user,
			"id", id,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	id, err = mw.next.Create(ctx, user)
	return
}

func (mw *loggingMiddleware) FindByID(ctx context.Context, id string) (user *model.User, err error) {
	l := log.WithPrefix(mw.logger, "method", "FindByID")
	level.Debug(l).Log(
		"id", id,
		"msg", "about to find a user by id",
	)

	defer func(begin time.Time) {
		level.Info(l).Log(
			"id", id,
			"user", user,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	user, err = mw.next.FindByID(ctx, id)
	return
}

func (mw *loggingMiddleware) FindByEMail(ctx context.Context, email string) (user *model.User, err error) {
	l := log.WithPrefix(mw.logger, "method", "FindByID")
	level.Debug(l).Log(
		"email", email,
		"msg", "about to find a user by email",
	)

	defer func(begin time.Time) {
		level.Info(l).Log(
			"email", email,
			"user", user,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	user, err = mw.next.FindByEMail(ctx, email)
	return
}

func (mw *loggingMiddleware) HasUsersWithRole(ctx context.Context, role model.Role) (exist bool, err error) {
	l := log.WithPrefix(mw.logger, "method", "HasUserWithRole")
	level.Debug(l).Log(
		"role", role,
		"msg", "about to determine if any users in a particular role do exist",
	)

	defer func() {
		level.Info(l).Log(
			"role", role,
			"exist", exist,
			"err", err,
		)
	}()

	exist, err = mw.next.HasUsersWithRole(ctx, role)
	return
}

func (mw *loggingMiddleware) clear(ctx context.Context) (count int64, err error) {
	l := log.WithPrefix(mw.logger, "method", "clear")
	level.Debug(l).Log("msg", "about to delete all users")

	defer func() {
		level.Info(l).Log(
			"err", err,
			"count", count,
		)
	}()

	count, err = mw.next.clear(ctx)
	return
}
