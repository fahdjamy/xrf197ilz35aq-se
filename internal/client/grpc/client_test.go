package grpc

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"testing"
	v1 "xrf197ilz35aq/gen/account/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const testGrpcAddress = "bufnet-test-conn"

var testLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))

type mockAccountServiceServer struct {
	v1.UnimplementedAccountServiceServer
}

func (server *mockAccountServiceServer) FindWallet(context.Context, *v1.FindWalletRequest) (*v1.FindWalletResponse, error) {
	fmt.Println("FindWallet called")
	return &v1.FindWalletResponse{}, nil
}

// newTestDialer creates a dialer function that connects to an in-memory gRPC server.
func newRPCTestDialer() RPCClientDialer {
	// a fake (simulated), in-memory network connection
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	v1.RegisterAccountServiceServer(server, &mockAccountServiceServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Mock dialer function we will inject into our ConnectionManager.
	return func(address string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		dialOpts := append(opts,
			// grpc.WithContextDialer option is a lets you override gRPC's default connection behavior.
			// removing grpc.WithContextDialer reverts to its default behavior, which is to perform a real network call
			grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
				return listener.Dial()
			}),
			// Sets up the security protocol to use over that connection
			grpc.WithTransportCredentials(insecure.NewCredentials()))

		return grpc.NewClient(address, dialOpts...)
	}
}

func TestConnectionManager_CreateOrGetConnection(t *testing.T) {
	manager := NewConnectionManager(newRPCTestDialer())
	defer manager.Close()

	// 1. first call should create connection
	conn1, err := manager.CreateOrGetConnection(testGrpcAddress, *testLogger)
	if err != nil {
		t.Fatalf("CreateOrGetConnection() failed: %v", err)
	}

	//2. second call should re-use the already created connection
	conn2, err := manager.CreateOrGetConnection(testGrpcAddress, slog.Logger{})
	if err != nil {
		t.Fatalf("CreateOrGetConnection() failed: %v", err)
	}

	if conn1 != conn2 {
		t.Fatalf("CreateOrGetConnection() returned different connections")
	}
}

func TestConnectionManager_CloseAll(t *testing.T) {
	manager := NewConnectionManager(newRPCTestDialer())
	defer manager.Close()

	conn1, err := manager.CreateOrGetConnection(testGrpcAddress, *testLogger)
	if err != nil {
		t.Fatalf("CreateOrGetConnection() failed: %v", err)
	}

	manager.CloseAll(*testLogger)

	if conn1.GetState() != connectivity.Shutdown {
		t.Fatalf("CloseAll() did not shutdown correctly")
	}
}

func TestConnectionManager_CloseConnection(t *testing.T) {
	manager := NewConnectionManager(newRPCTestDialer())
	defer manager.Close()

	conn1, err := manager.CreateOrGetConnection(testGrpcAddress, *testLogger)
	if err != nil {
		t.Fatalf("CreateOrGetConnection() failed: %v", err)
	}

	manager.CloseConnection(testGrpcAddress, *testLogger)
	if conn1.GetState() != connectivity.Shutdown {
		t.Fatalf("CloseConnection() did not shutdown correctly")
	}
}

func TestConnectionManager_Close(t *testing.T) {
	manager := NewConnectionManager(newRPCTestDialer())
	defer manager.Close()

	conn1, err := manager.CreateOrGetConnection(testGrpcAddress, *testLogger)
	if err != nil {
		t.Fatalf("CreateOrGetConnection() failed: %v", err)
	}
	manager.Close()

	if conn1.GetState() != connectivity.Shutdown {
		t.Fatalf("Close() did not shutdown correctly")
	}
}
