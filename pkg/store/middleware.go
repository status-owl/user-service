package store

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/status-owl/user-service/pkg/model"
)

// contains logging middleware for the UserStore

type Middleware func(UserStore) UserStore

func LoggingMiddleware(logger zerolog.Logger) Middleware {
	return func(next UserStore) UserStore {
		return &loggingMiddleware{
			logger: logger.With().
				Str("interface", "UserStore").
				Logger(),
			next: next,
		}
	}
}

type loggingMiddleware struct {
	logger zerolog.Logger
	next   UserStore
}

func (mw *loggingMiddleware) Delete(ctx context.Context, id string) (err error) {
	logger := mw.logger.With().Str("method", "Delete").Logger()

	logger.Trace().
		Str("id", id).
		Msg("about to delete an user")

	defer func() {
		if err != nil {
			logger.Error().
				Str("id", id).
				Err(err).
				Msg("failed to delete an user")
		} else {
			logger.Info().
				Str("id", id).
				Msg("user created")
		}
	}()

	err = mw.next.Delete(ctx, id)
	return
}

func (mw *loggingMiddleware) Create(ctx context.Context, user *model.User) (id string, err error) {
	logger := mw.logger.With().
		Str("method", "Create").
		Stringer("user", user).
		Logger()

	logger.Trace().
		Msg("about to create a user")

	defer func(begin time.Time) {
		logger = logger.With().
			Dur("took", time.Since(begin)).
			Logger()

		if err != nil {
			logger.Error().
				Err(err).
				Msg("failed to create an user")
		} else {
			logger.Info().
				Str("id", id).
				Msg("user created")
		}
	}(time.Now())

	id, err = mw.next.Create(ctx, user)
	return
}

func (mw *loggingMiddleware) FindByID(ctx context.Context, id string) (user *model.User, err error) {
	logger := mw.logger.With().
		Str("method", "FindByID").
		Str("id", id).
		Logger()

	logger.Trace().
		Msg("about to find a user by id")

	defer func(begin time.Time) {
		logger = logger.With().
			Dur("took", time.Since(begin)).
			Logger()

		if err != nil {
			logger.Error().
				Err(err).
				Msg("failed to find user")
		} else {
			logger.Info().
				Stringer("user", user).
				Msg("user found")
		}
	}(time.Now())

	user, err = mw.next.FindByID(ctx, id)
	return
}

func (mw *loggingMiddleware) FindByEMail(ctx context.Context, email string) (user *model.User, err error) {
	logger := mw.logger.With().
		Str("method", "FindByEmail").
		Logger()

	logger.Trace().
		Msg("about to find a user by email")

	defer func(begin time.Time) {
		logger = logger.With().
			Dur("took", time.Since(begin)).
			Logger()

		if err != nil {
			logger.Error().
				Err(err).
				Str("email", email).
				Msg("user not found")
		} else {
			logger.Info().
				Stringer("user", user).
				Msg("user found")
		}
	}(time.Now())

	user, err = mw.next.FindByEMail(ctx, email)
	return
}

func (mw *loggingMiddleware) HasUsersWithRole(ctx context.Context, role model.Role) (exist bool, err error) {
	logger := mw.logger.With().
		Str("method", "HasUserWithRole").
		Stringer("role", role).
		Logger()

	logger.Trace().
		Msg("about to determine if any users in a particular role do exist")

	defer func(begin time.Time) {
		logger = logger.With().
			Dur("took", time.Since(begin)).
			Logger()

		if err != nil {
			logger.Error().
				Err(err).
				Msg("failed to determine if any users in a particular role do exist")
		} else {
			logger.Info().
				Bool("exist", exist).
				Msg("determined if at least one user in a particular role does exist")
		}
	}(time.Now())

	exist, err = mw.next.HasUsersWithRole(ctx, role)
	return
}

func (mw *loggingMiddleware) clear(ctx context.Context) (count int64, err error) {
	logger := mw.logger.With().
		Str("method", "clear").
		Logger()

	logger.Trace().Msg("about to remove all users")

	defer func(begin time.Time) {
		logger = logger.With().
			Dur("took", time.Since(begin)).
			Logger()

		if err != nil {
			logger.Error().
				Err(err).
				Msg("failed to clear the db")
		} else {
			logger.Info().
				Int64("count", count).
				Msg("cleared db")
		}

	}(time.Now())

	count, err = mw.next.clear(ctx)
	return
}
