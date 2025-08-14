package grpc

import (
	"fmt"
	"log/slog"
	"sync"
	v1 "xrf197ilz35aq/proto/gen/proto/account/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectionManager handles the lifecycle of gRPC client connections.
// Ensures connections are reused and re-established when needed.
type ConnectionManager struct {
	// A mutex is used to protect the conns map during read/write operations to prevent race conditions.
	mut sync.RWMutex
	// connections stores active gRPC connections, keyed by server address. || sync.Map is a thread-safe map
	connections sync.Map
}

func (m *ConnectionManager) NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{}
}

func (m *ConnectionManager) CreateOrGetConnection(address string, logger slog.Logger) (*grpc.ClientConn, error) {
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
	// --- 2. Create a new connection if one doesn't exist or was closed ---
	newConn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("error creating gRPC connection: %w", err)
	}

	//////// 3. Store the new connection for future reuse ---
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

type XRFRPCServices struct {
	AccountClient v1.AccountServiceClient
}

func RegisterXRFRPCServices(address string, log slog.Logger, manager *ConnectionManager) (*XRFRPCServices, error) {
	conn, err := manager.CreateOrGetConnection(address, log)
	if err != nil {
		return nil, err
	}

	accountClient := v1.NewAccountServiceClient(conn)
	return &XRFRPCServices{
		AccountClient: accountClient,
	}, nil
}
