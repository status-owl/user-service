package transport

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/status-owl/user-service/pb"
	"github.com/status-owl/user-service/pkg/service"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	tests := []struct {
		name string
		// UserService return parameters
		id  string
		err error
		// want
		wantReply *pb.CreateUserReply
		wantErr   error
	}{
		{
			name:      "should return user id on success",
			id:        "123",
			err:       nil,
			wantReply: &pb.CreateUserReply{Id: "123"},
			wantErr:   nil,
		},
		{
			name: "should return an InvalidArgument error",
			id:   "123",
			err: &service.ValidationErrors{Errors: []service.ValidationError{
				{
					Name:   "email",
					Reason: "invalid",
				},
			}},
			wantReply: nil,
			wantErr:   grpcBadRequest("couldn't create a user due to invalid arguments", map[string]string{"email": "invalid"}),
		},
		{
			name:      "should return an AlreadyExists error if email is already in use",
			id:        "",
			err:       service.ErrEmailInUse,
			wantReply: nil,
			wantErr:   status.New(codes.AlreadyExists, "user with this email address already exists").Err(),
		},
		{
			name:      "should return an Internal error in case of unknown server errors",
			id:        "123",
			err:       errors.New("something went wrong"),
			wantReply: nil,
			wantErr:   status.Error(codes.Internal, "something went wrong"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, svc := setUpTest(t)

			svc.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(tt.id, tt.err)

			gotReply, gotErr := client.CreateUser(context.Background(), &pb.CreateUserRequest{
				Name:  "a",
				Email: "b",
			})

			a := assert.New(t)
			if tt.wantReply == nil {
				a.Nil(gotReply)
			} else {
				a.Equal(tt.wantReply.Id, gotReply.Id)
			}
			a.Equal(tt.wantErr, gotErr)
		})
	}
}

// grpcBadRequest creates a error with code=InvalidArgument and
// field violations
func grpcBadRequest(msg string, violations map[string]string) error {
	var fieldViolations []*errdetails.BadRequest_FieldViolation
	for field, description := range violations {
		fieldViolations = append(fieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: description,
		})
	}

	badRequest := errdetails.BadRequest{FieldViolations: fieldViolations}

	stat, err := status.New(codes.InvalidArgument, msg).WithDetails(&badRequest)
	if err != nil {
		panic("didn't expect an error, check your code!")
	}

	return stat.Err()
}

func setUpTest(t *testing.T) (pb.UserServiceClient, *service.MockUserService) {
	ctrl := gomock.NewController(t)
	svc := service.NewMockUserService(ctrl)

	var lis = bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	pb.RegisterUserServiceServer(srv, NewBaseGrpcServer(svc))

	go func() {
		if err := srv.Serve(lis); err != nil {
			panic("failed to start grpc server")
		}
	}()

	bufDialer := func(ctx context.Context, s string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.Dial("bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to the grpc server")
	}

	client := pb.NewUserServiceClient(conn)

	// do after the test is finished
	t.Cleanup(func() {
		ctrl.Finish()
		srv.Stop()
		_ = conn.Close()
	})

	return client, svc
}
