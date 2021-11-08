package store

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/status-owl/user-service/pkg/model"
)

// contains logging middleware for the UserStore

type loggingUserStore struct {
	logger log.Logger
	next   UserStore
}

func NewLoggingUserStore(next UserStore, logger log.Logger) UserStore {
	return &loggingUserStore{
		logger: log.WithPrefix(logger, "interface", "UserStore"),
		next:   next,
	}
}

func (s *loggingUserStore) Create(ctx context.Context, user *model.User) (id string, err error) {
	l := log.WithPrefix(s.logger, "method", "Create")
	level.Debug(l).Log("user", user, "msg", "about to create a user")

	defer func(begin time.Time) {
		level.Info(l).Log(
			"user", user,
			"id", id,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	id, err = s.next.Create(ctx, user)
	return
}

func (s *loggingUserStore) FindByID(ctx context.Context, id string) (user *model.User, err error) {
	l := log.WithPrefix(s.logger, "method", "FindByID")
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

	user, err = s.next.FindByID(ctx, id)
	return
}
