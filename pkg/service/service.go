package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/status-owl/user-service/pkg/model"
	"github.com/status-owl/user-service/pkg/store"
)

type UserService interface {
	Create(ctx context.Context, user *model.RequestedUser) (string, error)
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*model.User, error)
}

//go:generate mockgen -source service.go -destination mock.go -package $GOPACKAGE

var (
	ErrEmailInUse   = errors.New("user with requested email address already exists")
	ErrUserNotFound = errors.New("user not found")
)

type ValidationError struct {
	Name, Reason string
}

// Error satisfies error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid parameter %q: %s", e.Name, e.Reason)
}

type ValidationErrors struct {
	Errors []ValidationError
}

func (errors *ValidationErrors) Append(e ValidationError) ValidationErrors {
	return ValidationErrors{
		Errors: append(errors.Errors, e),
	}
}

// Error satisfies error interface
func (errors *ValidationErrors) Error() string {
	if len(errors.Errors) == 0 {
		return "unknown error"
	}

	var s string = errors.Errors[0].Error()
	for i := 1; i < len(errors.Errors); i++ {
		s = "\n" + errors.Errors[i].Error()
	}

	return s
}

func NewService(
	store store.UserStore,
	logger zerolog.Logger,
	js nats.JetStream,
) UserService {
	var svc UserService
	{
		svc = &userService{store}
		svc = EventingMiddleware(js)(svc)
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware()(svc)
	}
	return svc
}

type userService struct {
	userStore store.UserStore
}

func (s *userService) Delete(ctx context.Context, id string) error {
	return s.userStore.Delete(ctx, id)
}

func (s *userService) Create(ctx context.Context, user *model.RequestedUser) (string, error) {
	if err := s.validateRequestedUser(user); err != nil {
		return "", err
	}

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

func (s *userService) validateRequestedUser(user *model.RequestedUser) *ValidationErrors {
	var err ValidationErrors
	if len(strings.TrimSpace(user.EMail)) < 5 {
		err = err.Append(ValidationError{
			Name:   "email",
			Reason: "invalid email address",
		})
	}

	if len(bytes.TrimSpace(user.Pwd)) < 12 {
		err = err.Append(ValidationError{
			Name:   "password",
			Reason: "consider to use a stronger password, at least 12 characters long",
		})
	}

	if len(strings.TrimSpace(user.Name)) == 0 {
		err = err.Append(ValidationError{
			Name:   "name",
			Reason: "name is not set",
		})
	}

	// no validation errors
	if len(err.Errors) == 0 {
		return nil
	}

	return &err
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
	return s.userStore.FindByID(ctx, id)
}
