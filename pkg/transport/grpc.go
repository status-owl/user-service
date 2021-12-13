package transport

import "github.com/status-owl/user-service/pb"

//go:generate protoc --go_out=../../pb --go-grpc_out=../../pb  --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative  --proto_path=../../pb ../../pb/usersvc.proto
//go:generate mockgen -source ../../pb/usersvc_grpc.pb.go -destination ../../pb/usersvc_grpc_mock.pb.go -package pb

type grpcServer struct {
	pb.UnimplementedUserServiceServer
}
