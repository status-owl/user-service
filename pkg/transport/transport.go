package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/status-owl/user-service/pkg/api"
	"github.com/status-owl/user-service/pkg/endpoint"
	"github.com/status-owl/user-service/pkg/service"
)

func NewHTTPHandler(
	endpoints endpoint.Endpoints,
	tracer stdopentracing.Tracer,
	logger log.Logger,
) http.Handler {
	options := []httptransport.ServerOption{}

	m := mux.NewRouter()

	m.Handle("/users", httptransport.NewServer(
		endpoints.CreateUserEndpoint,
		decodeHTTPCreateUserRequest,
		encodeHTTPCreateUserResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(tracer, "CreateUser", logger)))...,
	)).Methods("POST")

	return m
}

func decodeHTTPCreateUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req api.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func encodeHTTPCreateUserResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	r := response.(endpoint.CreateUserResponse)

	var httpStatus int
	var httpResponse interface{}

	if r.Err != nil {
		if errors.Is(r.Err, service.ErrEmailInUse) {
			httpStatus = http.StatusBadRequest
			httpResponse = api.Problem{
				Status: http.StatusBadRequest,
				Title:  "bad request",
				InvalidParams: &[]api.InvalidParam{
					{Name: "email", Reason: "user with this email address already exists"},
				},
			}
		} else {
			httpStatus = http.StatusInternalServerError
			detail := r.Err.Error()
			httpResponse = api.Problem{
				Status: http.StatusInternalServerError,
				Title:  "internal server error",
				Detail: &detail,
			}
		}
	} else {
		httpStatus = http.StatusCreated
		httpResponse = api.CreateUserResponse{Id: r.ID}
	}

	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(httpStatus)
	return json.NewEncoder(w).Encode(httpResponse)
}
