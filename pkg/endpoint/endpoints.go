package endpoint

import (
	"context"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"github.com/status-owl/user-service/pkg/api"
	"golang.org/x/time/rate"

	"github.com/status-owl/user-service/pkg/service"
)

type Endpoints struct {
	CreateUserEndpoint endpoint.Endpoint
}

func NewEndpoints(
	svc service.UserService,
	logger log.Logger,
	duration metrics.Histogram,
	tracer stdopentracing.Tracer,
) Endpoints {
	var createUserEndpoint endpoint.Endpoint
	{
		createUserEndpoint = MakeCreateUserEndpoint(svc)
		createUserEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(createUserEndpoint)
		createUserEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createUserEndpoint)
		createUserEndpoint = opentracing.TraceServer(tracer, "CreateUser")(createUserEndpoint)
		createUserEndpoint = LoggingMiddleware(log.With(logger, "method", "CreateUser"))(createUserEndpoint)
		createUserEndpoint = InstumentingMiddleware(duration.With("method", "CreateUser"))(createUserEndpoint)
	}

	return Endpoints{
		CreateUserEndpoint: createUserEndpoint,
	}
}

func MakeCreateUserEndpoint(s service.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.CreateUserRequest)
		id, err := s.Create(ctx, &service.RequestedUser{
			EMail: req.Email,
			Name:  req.Name,
			Pwd:   []byte(req.Password),
		})
		return CreateUserResponse{ID: id, Err: err}, nil
	}
}

// compile time assertions for our response type implementing endpoint.Failer
var _ endpoint.Failer = CreateUserResponse{}

type CreateUserRequest struct {
}

type CreateUserResponse struct {
	ID  string
	Err error
}

func (r CreateUserResponse) Failed() error {
	return r.Err
}
