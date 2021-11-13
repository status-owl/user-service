package service

import (
	"context"
	"errors"
	"time"

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

const storeTimeout = 5 * time.Second

var (
	ErrEmailInUse = errors.New("user with requested email address already exists")
)

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

func (s *userService) initialize() error {
	_, cancel := context.WithTimeout(context.Background(), storeTimeout)
	defer cancel()

	// 1. look for admin users
	// 2. if there are not any - create one with random password and log it
	return nil
}

func (s *userService) Create(ctx context.Context, user *RequestedUser) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()

	// check if the user with given email already exists
	userExists, err := s.hasUserWithEMail(ctx, user.EMail)
	if err != nil {
		return "", err
	}

	if userExists {
		return "", ErrEmailInUse
	}

	return s.userStore.Create(ctx, &model.User{Name: user.Name, EMail: user.EMail})
}

func (s *userService) hasUserWithEMail(ctx context.Context, email string) (bool, error) {
	_, err := s.userStore.FindByEMail(ctx, email)
	if err != nil {
		if err == store.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *userService) FindByID(ctx context.Context, id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()

	return s.userStore.FindByID(ctx, id)
}
