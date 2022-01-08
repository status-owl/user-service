package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"net/http"
	"time"

	"github.com/status-owl/user-service/pkg/service"
)

//go:generate oapi-codegen -o model.go --generate=types --package=$GOPACKAGE ../../spec/api-v1.yaml

// NewHTTPHandler creates and returns a configured http.Handler
func NewHTTPHandler(svc service.UserService, logger zerolog.Logger) http.Handler {
	return LoggingMiddleware(logger, NewBaseHTTPHandler(svc))
}

// NewBaseHTTPHandler returns a base http.Handler without any configured middlewares
func NewBaseHTTPHandler(svc service.UserService) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/users/", findUserByID(svc))
	return mux
}

func findUserByID(svc service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		id := r.URL.Path[len("/users/"):]

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		user, err := svc.FindByID(ctx, id)
		if err != nil {
			handleError(w, err2Problem(err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(&User{
			Email: user.EMail,
			Id:    user.ID,
			Name:  user.Name,
		}); err != nil {
			panic("failed to encode json")
		}
	}
}

func decodeRequest(r *http.Request, v interface{}) (*Problem, error) {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		p := Problem{
			Status: http.StatusBadRequest,
			Title:  http.StatusText(http.StatusBadRequest),
			Detail: fmt.Sprintf("unexpected payload, expected JSON: %s", err.Error()),
		}
		return &p, err
	}

	return nil, nil
}

func handleError(w http.ResponseWriter, p *Problem) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(int(p.Status))

	err := json.NewEncoder(w).Encode(p)
	if err != nil {
		panic(err)
	}
}

func err2Problem(err error) *Problem {
	if err == nil {
		panic("given error supposed not to be nil")
	}

	var p Problem

	// first check for validation errors
	if verr, ok := err.(*service.ValidationErrors); ok {
		var params []InvalidParam
		for _, ve := range verr.Errors {
			params = append(params, InvalidParam{
				Name:   ve.Name,
				Reason: ve.Reason,
			})
		}
		p = Problem{
			Status:        http.StatusBadRequest,
			Title:         http.StatusText(http.StatusBadRequest),
			Detail:        "One of the parameters is invalid",
			InvalidParams: &params,
		}
	} else if errors.Is(err, service.ErrEmailInUse) {
		p = Problem{
			Status: http.StatusBadRequest,
			Title:  http.StatusText(http.StatusBadRequest),
			InvalidParams: &[]InvalidParam{
				{Name: "email", Reason: "user with this email address already exists"},
			},
		}
	} else if errors.Is(err, service.ErrUserNotFound) {
		p = Problem{
			Status: http.StatusNotFound,
			Title:  http.StatusText(http.StatusNotFound),
			Detail: "user with given id doesn't exist",
		}
	} else {
		p = Problem{
			Status: http.StatusInternalServerError,
			Title:  http.StatusText(http.StatusInternalServerError),
			Detail: err.Error(),
		}
	}

	return &p
}
