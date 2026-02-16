package httpapi

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	registryv1 "github.com/onlyboxes/onlyboxes/api/gen/go/registry/v1"
	"github.com/onlyboxes/onlyboxes/console/internal/grpcserver"
	"github.com/onlyboxes/onlyboxes/console/internal/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestRegisterAndListLifecycle(t *testing.T) {
	store := registry.NewStore()
	registrySvc := grpcserver.NewRegistryService(store, 5, 15)
	grpcSrv := grpcserver.NewServer("test-token", registrySvc)
	grpcListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to open grpc listener: %v", err)
	}
	defer grpcListener.Close()
	go func() {
		_ = grpcSrv.Serve(grpcListener)
	}()
	defer grpcSrv.Stop()

	handler := NewWorkerHandler(store, 15*time.Second)
	router := NewRouter(handler)
	httpSrv := httptest.NewServer(router)
	defer httpSrv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(grpcListener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial grpc: %v", err)
	}
	defer conn.Close()

	client := registryv1.NewWorkerRegistryServiceClient(conn)
	authCtx := metadata.AppendToOutgoingContext(ctx, grpcserver.HeaderSharedToken, "test-token")
	_, err = client.Register(authCtx, &registryv1.RegisterRequest{
		NodeId:       "node-int-1",
		NodeName:     "integration-node",
		ExecutorKind: "docker",
		Languages: []*registryv1.LanguageCapability{{
			Language: "python",
			Version:  "3.12",
		}},
		Version: "v0.1.0",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	listOnline := requestList(t, httpSrv.URL+"/api/v1/workers?status=online")
	if listOnline.Total != 1 || len(listOnline.Items) != 1 {
		t.Fatalf("expected one online worker after register, got total=%d len=%d", listOnline.Total, len(listOnline.Items))
	}
	if listOnline.Items[0].NodeID != "node-int-1" || listOnline.Items[0].Status != registry.StatusOnline {
		t.Fatalf("unexpected online worker payload")
	}

	handler.nowFn = func() time.Time {
		return time.Now().Add(16 * time.Second)
	}
	listOffline := requestList(t, httpSrv.URL+"/api/v1/workers?status=offline")
	if listOffline.Total != 1 || len(listOffline.Items) != 1 {
		t.Fatalf("expected one offline worker after heartbeat timeout, got total=%d len=%d", listOffline.Total, len(listOffline.Items))
	}
	if listOffline.Items[0].NodeID != "node-int-1" || listOffline.Items[0].Status != registry.StatusOffline {
		t.Fatalf("unexpected offline worker payload")
	}
}

func requestList(t *testing.T, url string) listWorkersResponse {
	t.Helper()

	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to call list API: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 response, got %d", res.StatusCode)
	}

	var payload listWorkersResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	return payload
}
