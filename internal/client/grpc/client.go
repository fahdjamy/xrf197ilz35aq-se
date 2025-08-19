package grpc

import (
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

type RPCClientDialer func(address string, opts ...grpc.DialOption) (*grpc.ClientConn, error)

// DialOptionProvider defines a function that returns the necessary dial options for a connection.
type DialOptionProvider func(address string) ([]grpc.DialOption, error)

// ConnectionManager handles the lifecycle of gRPC client connections.
// Ensures connections are reused and re-established when needed.
type ConnectionManager struct {
	// A mutex is used to protect the conns map during read/write operations to prevent race conditions.
	mut sync.RWMutex
	// connections stores active gRPC connections, keyed by server address. || sync.Map is a thread-safe map
	connections sync.Map
	dialer      RPCClientDialer
	dialOptions DialOptionProvider
}

func NewConnectionManager(dialer RPCClientDialer, optionProvider DialOptionProvider) *ConnectionManager {
	if dialer == nil {
		// default to gRPC NewClient dialer
		dialer = grpc.NewClient
	}
	return &ConnectionManager{
		dialer:      dialer,
		dialOptions: optionProvider,
	}
}

func (m *ConnectionManager) CreateOrGetConnection(address string, logger slog.Logger, certPath string) (*grpc.ClientConn, error) {
	// Ensure that the connection-checking and creation logic is atomic.
	m.mut.RLock()
	defer m.mut.RUnlock()

	//////// 1. Check for an existing, healthy connection ---
	if conn, ok := m.connections.Load(address); ok && conn != nil {
		clientConn := conn.(*grpc.ClientConn)
		if clientConn.GetState() != connectivity.Shutdown {
			return clientConn, nil
		}
		logger.Info("connection was already closed", "address", address)
	}

	logger.Info("creating new gRPC connection", "address", address)

	//////// 2. Get the required dial options
	opts, err := m.dialOptions(address)
	if err != nil {
		return nil, err
	}

	//////// 3. Create a new connection if one doesn't exist or was closed ---

	newConn, err := m.dialer(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating gRPC connection: %w", err)
	}

	//////// 4. Store the new connection for future reuse ---
	m.connections.Store(address, newConn)
	return newConn, nil
}

func (m *ConnectionManager) CloseConnection(address string, log slog.Logger) {
	m.mut.Lock()
	defer m.mut.Unlock()

	if conn, ok := m.connections.LoadAndDelete(address); ok && conn != nil {
		err := conn.(*grpc.ClientConn).Close()
		if err != nil {
			log.Warn("error closing gRPC connection", "address", address, "err", err)
		}
		log.Info("closed gRPC connection", "address", address)
		m.connections.Delete(address)
	}
}

func (m *ConnectionManager) CloseAll(log slog.Logger) {
	m.mut.Lock()
	defer m.mut.Unlock()
	log.Info("closing all gRPC connections")
	m.connections.Range(func(key, value interface{}) bool {
		if conn, ok := m.connections.LoadAndDelete(key); ok && conn != nil {
			err := conn.(*grpc.ClientConn).Close()
			if err != nil {
				log.Warn("error closing gRPC connection", "address", key, "err", err)
			}
		}
		return true
	})
}

// Close is used to close all connections but ideally, user should call CloseAll (ConnectionManager.CloseAll)
func (m *ConnectionManager) Close() {
	m.mut.Lock()
	defer m.mut.Unlock()
	m.connections.Range(func(key, value interface{}) bool {
		if conn, ok := m.connections.LoadAndDelete(key); ok && conn != nil {
			_ = conn.(*grpc.ClientConn).Close()
		}
		return true
	})
}

// NewTLSDialOptionProvider creates a provider that configures a client for TLS.
func NewTLSDialOptionProvider(certPath, serverName string) DialOptionProvider {
	return func(address string) ([]grpc.DialOption, error) {
		caCert, err := os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("could not read certificate: %w", err)
		}

		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("could not append certificate to pool")
		}

		creds := credentials.NewClientTLSFromCert(certPool, serverName)
		return []grpc.DialOption{grpc.WithTransportCredentials(creds)}, nil
	}
}
