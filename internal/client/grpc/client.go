package grpc

import (
	"fmt"
	"xrf197ilz35aq/internal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For dev/testing only
)

type Client struct {
	config internal.GrpcConfig
	conn   *grpc.ClientConn
}

func (c *Client) Client() (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		c.config.Address,
		// For production, always use secure credentials!
		// e.g., grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc.NewClient: %w", err)
	}
	c.conn = conn
	return conn, nil
}

func (c *Client) Register(serviceRegisters ...func(conn *grpc.ClientConn)) error {
	if c.conn == nil {
		return fmt.Errorf("conn is nil")
	}
	for _, register := range serviceRegisters {
		register(c.conn)
	}
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
