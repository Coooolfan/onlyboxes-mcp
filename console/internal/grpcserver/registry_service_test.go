package grpcserver

import (
	"context"
	"net"
	"strings"
	"testing"

	registryv1 "github.com/onlyboxes/onlyboxes/api/gen/go/registry/v1"
	"github.com/onlyboxes/onlyboxes/console/internal/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func newBufClient(t *testing.T, token string, svc registryv1.WorkerRegistryServiceServer) (registryv1.WorkerRegistryServiceClient, func()) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.UnaryInterceptor(UnaryTokenAuthInterceptor(token)))
	registryv1.RegisterWorkerRegistryServiceServer(server, svc)
	go func() {
		_ = server.Serve(listener)
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.NewClient(
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		server.Stop()
		_ = listener.Close()
	}
	return registryv1.NewWorkerRegistryServiceClient(conn), cleanup
}

func withToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, HeaderSharedToken, token)
}

func TestRegisterAuthRejectsMissingToken(t *testing.T) {
	store := registry.NewStore()
	svc := NewRegistryService(store, 5, 15)
	client, cleanup := newBufClient(t, "test-token", svc)
	defer cleanup()

	_, err := client.Register(context.Background(), &registryv1.RegisterRequest{NodeId: "node-1"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
}

func TestHeartbeatReturnsNotFoundForUnknownNode(t *testing.T) {
	store := registry.NewStore()
	svc := NewRegistryService(store, 5, 15)
	client, cleanup := newBufClient(t, "test-token", svc)
	defer cleanup()

	ctx := withToken(context.Background(), "test-token")
	_, err := client.Heartbeat(ctx, &registryv1.HeartbeatRequest{NodeId: "missing-node"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestRegisterReturnsExpectedResponseFields(t *testing.T) {
	store := registry.NewStore()
	svc := NewRegistryService(store, 5, 15)
	client, cleanup := newBufClient(t, "test-token", svc)
	defer cleanup()

	ctx := withToken(context.Background(), "test-token")
	resp, err := client.Register(ctx, &registryv1.RegisterRequest{NodeId: "node-1", NodeName: "n1"})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if resp.GetServerTimeUnixMs() <= 0 {
		t.Fatalf("expected server_time_unix_ms > 0")
	}
	if resp.GetHeartbeatIntervalSec() != 5 {
		t.Fatalf("expected heartbeat_interval_sec=5, got %d", resp.GetHeartbeatIntervalSec())
	}
	if resp.GetOfflineTtlSec() != 15 {
		t.Fatalf("expected offline_ttl_sec=15, got %d", resp.GetOfflineTtlSec())
	}
}

func TestRegisterAuthAcceptsTokenFromConfiguredList(t *testing.T) {
	store := registry.NewStore()
	svc := NewRegistryService(store, 5, 15)
	client, cleanup := newBufClient(t, "token-a, token-b", svc)
	defer cleanup()

	ctx := withToken(context.Background(), "token-b")
	_, err := client.Register(ctx, &registryv1.RegisterRequest{NodeId: "node-list-1", NodeName: "n-list"})
	if err != nil {
		t.Fatalf("expected token from list to pass auth, got %v", err)
	}
}

func TestRegisterAuthRejectsInvalidConfiguredTokenList(t *testing.T) {
	store := registry.NewStore()
	svc := NewRegistryService(store, 5, 15)
	client, cleanup := newBufClient(t, ",,", svc)
	defer cleanup()

	ctx := withToken(context.Background(), "any-token")
	_, err := client.Register(ctx, &registryv1.RegisterRequest{NodeId: "node-invalid-1", NodeName: "n-invalid"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated for invalid auth config, got %v", err)
	}
}

func TestRegisterRejectsNilRequest(t *testing.T) {
	svc := NewRegistryService(registry.NewStore(), 5, 15)
	_, err := svc.Register(context.Background(), nil)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestRegisterRejectsBlankNodeID(t *testing.T) {
	svc := NewRegistryService(registry.NewStore(), 5, 15)
	_, err := svc.Register(context.Background(), &registryv1.RegisterRequest{NodeId: "   "})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestRegisterRejectsTooLongNodeID(t *testing.T) {
	svc := NewRegistryService(registry.NewStore(), 5, 15)
	tooLongNodeID := strings.Repeat("n", maxNodeIDLength+1)
	_, err := svc.Register(context.Background(), &registryv1.RegisterRequest{NodeId: tooLongNodeID})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestHeartbeatRejectsTooLongNodeID(t *testing.T) {
	svc := NewRegistryService(registry.NewStore(), 5, 15)
	tooLongNodeID := strings.Repeat("h", maxNodeIDLength+1)
	_, err := svc.Heartbeat(context.Background(), &registryv1.HeartbeatRequest{NodeId: tooLongNodeID})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestRegisterRespectsCanceledContext(t *testing.T) {
	svc := NewRegistryService(registry.NewStore(), 5, 15)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Register(ctx, &registryv1.RegisterRequest{NodeId: "node-ctx-cancel"})
	if status.Code(err) != codes.Canceled {
		t.Fatalf("expected Canceled, got %v", err)
	}
}
