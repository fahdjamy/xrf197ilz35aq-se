package grpc

import (
	"fmt"
	"sync"
	"xrf197ilz35aq/internal"
	v1 "xrf197ilz35aq/proto/gen/proto/account/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For dev/testing only
)

var (
	// 1. Declare a sync.Once and a variable to hold our connection.
	once     sync.Once
	grpcConn *grpc.ClientConn
	connErr  error
)

type RPCServices struct {
	AccountClient v1.AccountServiceClient
}

type RPCClient struct {
	config internal.GrpcConfig
}

func (c *RPCClient) Connect(config internal.GrpcConfig) (*grpc.ClientConn, error) {
	return createConnection(config)
}

func createConnection(config internal.GrpcConfig) (*grpc.ClientConn, error) {
	once.Do(func() {
		c, err := grpc.NewClient(
			config.Address,
			// For production, always use secure credentials!
			// e.g., grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			connErr = err // Store the error if connection fails.
			return
		}
		grpcConn = c
	})
	if connErr != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", connErr)
	}

	return grpcConn, nil
}

func (c *RPCClient) Register() (*RPCServices, error) {
	if grpcConn == nil {
		return nil, fmt.Errorf("conn is nil")
	}

	acctClient := v1.NewAccountServiceClient(grpcConn)
	return &RPCServices{AccountClient: acctClient}, nil
}
