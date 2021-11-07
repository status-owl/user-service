package service

import (
	"context"
)

type UserService interface {
	Create(ctx context.Context, user *RequestedUser) (string, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type RequestedUser struct {
	EMail, Name string
	Pwd         []byte
}

type userService struct {
	userStore UserStore
}

func NewService(store UserStore) UserService {
	return &userService{userStore: store}
}

func (s *userService) Create(ctx context.Context, user *RequestedUser) (string, error) {
	return s.userStore.Create(ctx, &User{Name: user.Name, EMail: user.EMail})
}

func (s *userService) FindByID(ctx context.Context, id string) (*User, error) {
	return s.userStore.FindByID(ctx, id)
}

type Role string

const (
	Admin     Role = "ADMIN"
	Reporter  Role = "REPORTER"
	Undefined Role = "UNDEFINED"
)

func RoleFromString(s string) Role {
	switch s {
	case string(Admin):
		return Admin
	case string(Reporter):
		return Reporter
	default:
		return Undefined
	}
}

// User represents an application user
type User struct {
	ID      string
	Name    string
	EMail   string
	PwdHash string
	Role    Role
}
