package service

import (
	"context"

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

type userService struct {
	userStore store.UserStore
}

func NewService(store store.UserStore) UserService {
	return &userService{userStore: store}
}

func (s *userService) Create(ctx context.Context, user *RequestedUser) (string, error) {
	return s.userStore.Create(ctx, &model.User{Name: user.Name, EMail: user.EMail})
}

func (s *userService) FindByID(ctx context.Context, id string) (*model.User, error) {
	return s.userStore.FindByID(ctx, id)
}
