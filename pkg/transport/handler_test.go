package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/status-owl/user-service/pkg/model"
	"github.com/status-owl/user-service/pkg/service"
	"github.com/stretchr/testify/assert"
)

func TestFindUserByID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		// user service return parameters
		user *model.User
		err  error
		// want
		code     int
		response interface{}
	}{
		{
			name: "should respond with 200 for a valid user",
			id:   "123",
			user: &model.User{
				ID:    "123",
				Name:  "John",
				EMail: "john@example.com",
			},
			err:  nil,
			code: http.StatusOK,
			response: &User{
				Email: "john@example.com",
				Id:    "123",
				Name:  "John",
			},
		},
		{
			name: "should respond with 404 if no user is available",
			id:   "123",
			user: nil,
			err:  service.ErrUserNotFound,
			code: http.StatusNotFound,
			response: &Problem{
				Detail: "user with given id doesn't exist",
				Status: http.StatusNotFound,
				Title:  http.StatusText(http.StatusNotFound),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			b, err := json.Marshal(tt.user)
			a.Nil(err)

			req, err := http.NewRequest("GET", "/users/"+tt.id, bytes.NewReader(b))
			a.Nil(err)

			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := service.NewMockUserService(ctrl)

			svc.
				EXPECT().
				FindByID(gomock.Any(), gomock.Eq(tt.id)).
				Return(tt.user, tt.err)

			NewBaseHTTPHandler(svc).ServeHTTP(rr, req)

			a.Equal(tt.code, rr.Code)

			switch expectedResponse := tt.response.(type) {
			case *User:
				var actualResponse User
				err = json.NewDecoder(rr.Body).Decode(&actualResponse)
				a.Nil(err)

				a.Equal("application/json; charset=utf-8", rr.Header().Get("content-type"))
				a.Equal(*expectedResponse, actualResponse)

			case *Problem:
				var actualResponse Problem
				err = json.NewDecoder(rr.Body).Decode(&actualResponse)
				a.Nil(err)

				a.Equal("application/problem+json; charset=utf-8", rr.Header().Get("content-type"))
				a.Equal(*expectedResponse, actualResponse)
			default:
				panic("unexpected response type")
			}
		})
	}
}
