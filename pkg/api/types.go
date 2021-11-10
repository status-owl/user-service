// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.0 DO NOT EDIT.
package api

// CreateUserRequest defines model for CreateUserRequest.
type CreateUserRequest struct {
	// Email address
	Email string `json:"email"`

	// User name
	Name string `json:"name"`

	// Password
	Password string `json:"password"`
}

// CreateUserResponse defines model for CreateUserResponse.
type CreateUserResponse struct {
	Id string `json:"id"`
}

// Represents an invalid property in a bad request
type InvalidParam struct {
	// Name of the property
	Name string `json:"name"`

	// Why is the property considered invalid
	Reason string `json:"reason"`
}

// Problem defines model for Problem.
type Problem struct {
	// A human-readable explanation specific to this occurrence of the problem.
	Detail        *string         `json:"detail,omitempty"`
	InvalidParams *[]InvalidParam `json:"invalid-params,omitempty"`

	// The HTTP status code
	Status int `json:"status"`

	// A short, human-readable summary of the problem type
	Title string `json:"title"`

	// A URI reference that identifies the problem type
	Type *string `json:"type,omitempty"`
}

// CreateUserJSONBody defines parameters for CreateUser.
type CreateUserJSONBody CreateUserRequest

// CreateUserJSONRequestBody defines body for CreateUser for application/json ContentType.
type CreateUserJSONRequestBody CreateUserJSONBody
