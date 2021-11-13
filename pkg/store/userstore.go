package store

import (
	"context"
	"errors"
	"github.com/go-kit/log"
	"github.com/status-owl/user-service/pkg/model"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserStore is responsible for storing and fetching of users
type UserStore interface {
	Create(ctx context.Context, user *model.User) (string, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEMail(ctx context.Context, email string) (*model.User, error)
	HasUsersWithRole(ctx context.Context, role model.Role) (bool, error)
	clear(ctx context.Context) (int64, error)
}

var (
	//ErrNotFound signals that a user could not be found
	ErrNotFound = errors.New("user not found")
)

func NewUserStore(client *mongo.Client, logger log.Logger) (UserStore, error) {
	store := &mongoUserStore{client}
	if err := store.createIndexes(); err != nil {
		return nil, err
	}

	return LoggingMiddleware(logger)(store), nil
}
