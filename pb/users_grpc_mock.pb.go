// Code generated by MockGen. DO NOT EDIT.
// Source: ../../pb/users_grpc.pb.go

// Package pb is a generated GoMock package.
package pb

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockUsersClient is a mock of UsersClient interface.
type MockUsersClient struct {
	ctrl     *gomock.Controller
	recorder *MockUsersClientMockRecorder
}

// MockUsersClientMockRecorder is the mock recorder for MockUsersClient.
type MockUsersClientMockRecorder struct {
	mock *MockUsersClient
}

// NewMockUsersClient creates a new mock instance.
func NewMockUsersClient(ctrl *gomock.Controller) *MockUsersClient {
	mock := &MockUsersClient{ctrl: ctrl}
	mock.recorder = &MockUsersClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUsersClient) EXPECT() *MockUsersClientMockRecorder {
	return m.recorder
}

// CreateUser mocks base method.
func (m *MockUsersClient) CreateUser(ctx context.Context, in *CreateUserRequest, opts ...grpc.CallOption) (*CreateUserReply, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateUser", varargs...)
	ret0, _ := ret[0].(*CreateUserReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUsersClientMockRecorder) CreateUser(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUsersClient)(nil).CreateUser), varargs...)
}

// DeleteUser mocks base method.
func (m *MockUsersClient) DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserReply, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteUser", varargs...)
	ret0, _ := ret[0].(*DeleteUserReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUsersClientMockRecorder) DeleteUser(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUsersClient)(nil).DeleteUser), varargs...)
}

// UpdateUser mocks base method.
func (m *MockUsersClient) UpdateUser(ctx context.Context, in *UpdateUserRequest, opts ...grpc.CallOption) (*UpdateUserReply, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateUser", varargs...)
	ret0, _ := ret[0].(*UpdateUserReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockUsersClientMockRecorder) UpdateUser(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockUsersClient)(nil).UpdateUser), varargs...)
}

// MockUsersServer is a mock of UsersServer interface.
type MockUsersServer struct {
	ctrl     *gomock.Controller
	recorder *MockUsersServerMockRecorder
}

// MockUsersServerMockRecorder is the mock recorder for MockUsersServer.
type MockUsersServerMockRecorder struct {
	mock *MockUsersServer
}

// NewMockUsersServer creates a new mock instance.
func NewMockUsersServer(ctrl *gomock.Controller) *MockUsersServer {
	mock := &MockUsersServer{ctrl: ctrl}
	mock.recorder = &MockUsersServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUsersServer) EXPECT() *MockUsersServerMockRecorder {
	return m.recorder
}

// CreateUser mocks base method.
func (m *MockUsersServer) CreateUser(arg0 context.Context, arg1 *CreateUserRequest) (*CreateUserReply, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", arg0, arg1)
	ret0, _ := ret[0].(*CreateUserReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUsersServerMockRecorder) CreateUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUsersServer)(nil).CreateUser), arg0, arg1)
}

// DeleteUser mocks base method.
func (m *MockUsersServer) DeleteUser(arg0 context.Context, arg1 *DeleteUserRequest) (*DeleteUserReply, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", arg0, arg1)
	ret0, _ := ret[0].(*DeleteUserReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUsersServerMockRecorder) DeleteUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUsersServer)(nil).DeleteUser), arg0, arg1)
}

// UpdateUser mocks base method.
func (m *MockUsersServer) UpdateUser(arg0 context.Context, arg1 *UpdateUserRequest) (*UpdateUserReply, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", arg0, arg1)
	ret0, _ := ret[0].(*UpdateUserReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockUsersServerMockRecorder) UpdateUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockUsersServer)(nil).UpdateUser), arg0, arg1)
}

// mustEmbedUnimplementedUsersServer mocks base method.
func (m *MockUsersServer) mustEmbedUnimplementedUsersServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedUsersServer")
}

// mustEmbedUnimplementedUsersServer indicates an expected call of mustEmbedUnimplementedUsersServer.
func (mr *MockUsersServerMockRecorder) mustEmbedUnimplementedUsersServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedUsersServer", reflect.TypeOf((*MockUsersServer)(nil).mustEmbedUnimplementedUsersServer))
}

// MockUnsafeUsersServer is a mock of UnsafeUsersServer interface.
type MockUnsafeUsersServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeUsersServerMockRecorder
}

// MockUnsafeUsersServerMockRecorder is the mock recorder for MockUnsafeUsersServer.
type MockUnsafeUsersServerMockRecorder struct {
	mock *MockUnsafeUsersServer
}

// NewMockUnsafeUsersServer creates a new mock instance.
func NewMockUnsafeUsersServer(ctrl *gomock.Controller) *MockUnsafeUsersServer {
	mock := &MockUnsafeUsersServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeUsersServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnsafeUsersServer) EXPECT() *MockUnsafeUsersServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedUsersServer mocks base method.
func (m *MockUnsafeUsersServer) mustEmbedUnimplementedUsersServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedUsersServer")
}

// mustEmbedUnimplementedUsersServer indicates an expected call of mustEmbedUnimplementedUsersServer.
func (mr *MockUnsafeUsersServerMockRecorder) mustEmbedUnimplementedUsersServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedUsersServer", reflect.TypeOf((*MockUnsafeUsersServer)(nil).mustEmbedUnimplementedUsersServer))
}