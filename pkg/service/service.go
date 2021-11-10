package service

import (
	"context"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/log"
	"github.com/status-owl/user-service/pkg/model"
	"github.com/status-owl/user-service/pkg/store"
)

type UserService interface {
	Create(ctx context.Context, user *RequestedUser) (string, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
}

type RequestedUser struct {
	EMail, Name string
	Pwd         []byte
}

func NewService(store store.UserStore, logger log.Logger, createdUsers, fetchedUsers metrics.Counter) UserService {
	var svc UserService
	{
		svc = &userService{store}
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(createdUsers, fetchedUsers)(svc)
	}
	return svc
}

type userService struct {
	userStore store.UserStore
}

func (s *userService) Create(ctx context.Context, user *RequestedUser) (string, error) {
	return s.userStore.Create(ctx, &model.User{Name: user.Name, EMail: user.EMail})
}

func (s *userService) FindByID(ctx context.Context, id string) (*model.User, error) {
	return s.userStore.FindByID(ctx, id)
}
