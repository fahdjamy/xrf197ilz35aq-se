package internal

import (
	"errors"
	"fmt"
)

type ExternalError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *ExternalError) Error() string {
	return fmt.Sprintf("processing error: %s", e.Message)
}

type APIClientError struct {
	Message string `json:"error"`
	Code    int    `json:"code"`
}

func (e *APIClientError) Error() string {
	return fmt.Sprintf("api client error: %s, code: %d", e.Message, e.Code)
}

type ServerError struct {
	Message string `json:"message"`
	Err     error  `json:"error"`
}

func (aErr *ServerError) Error() string {
	return fmt.Sprintf("Internal Service error: %s", aErr.Message)
}

var ErrInvalidGRPCTimeStamp = errors.New("invalid from gRPC server")
var GrpcConnectionClosedErr = errors.New("grpc connection closed")
