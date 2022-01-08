package transport

import (
	"context"
	"errors"
	"github.com/status-owl/user-service/pb"
	"github.com/status-owl/user-service/pkg/model"
	"github.com/status-owl/user-service/pkg/service"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate protoc --go_out=../../pb --go-grpc_out=../../pb  --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative  --proto_path=../../pb ../../pb/usersvc.proto
//go:generate mockgen -source ../../pb/usersvc_grpc.pb.go -destination ../../pb/usersvc_grpc_mock.pb.go -package pb

type grpcServer struct {
	svc service.UserService
	pb.UnimplementedUserServiceServer
}

func NewBaseGrpcServer(svc service.UserService) pb.UserServiceServer {
	return &grpcServer{svc: svc}
}

func (s grpcServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	id, err := s.svc.Create(ctx, model.RequestedUser{
		EMail: req.Email,
		Name:  req.Name,
	})

	if err != nil {
		return nil, err2GrpcStatus(err).Err()
	}

	return &pb.CreateUserReply{Id: id}, nil
}

func err2GrpcStatus(err error) *status.Status {

	var stat *status.Status

	if verr, ok := err.(*service.ValidationErrors); ok {
		var fieldViolations []*errdetails.BadRequest_FieldViolation
		for _, ve := range verr.Errors {
			fieldViolations = append(fieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       ve.Name,
				Description: ve.Reason,
			})
		}

		stat, err = status.New(
			codes.InvalidArgument,
			"couldn't create a user due to invalid arguments",
		).WithDetails(&errdetails.BadRequest{FieldViolations: fieldViolations})

		if err != nil {
			panic("didn't except any errors, check your code!")
		}
	} else if errors.Is(err, service.ErrEmailInUse) {
		stat = status.New(codes.AlreadyExists, "user with this email address already exists")
	} else if errors.Is(err, service.ErrUserNotFound) {
		stat = status.New(codes.NotFound, "user with given id doesn't exist")
	} else {
		stat = status.New(codes.Internal, err.Error())
	}

	return stat
}
