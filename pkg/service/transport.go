package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/status-owl/user-service/pkg/api"
)

func decodeCreateUserRequest(_ context.Context, httpReq *http.Request) (interface{}, error) {
	var req api.CreateUserRequest
	err := json.NewDecoder(httpReq.Body).Decode(&req)
	return req, err
}
