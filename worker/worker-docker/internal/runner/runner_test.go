package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/onlyboxes/onlyboxes/api/pkg/registryauth"
	"github.com/onlyboxes/onlyboxes/worker/worker-docker/internal/config"
	"google.golang.org/grpc/metadata"
)

func TestRunReturnsContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Run(ctx, testConfig())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRunWaitsBeforeReconnectOnSessionFailure(t *testing.T) {
	originalWaitReconnect := waitReconnect
	waitCalls := 0
	waitReconnect = func(context.Context) error {
		waitCalls++
		return context.Canceled
	}
	defer func() {
		waitReconnect = originalWaitReconnect
	}()

	cfg := testConfig()
	cfg.ConsoleGRPCTarget = "127.0.0.1:1"
	cfg.CallTimeout = 5 * time.Millisecond

	err := Run(context.Background(), cfg)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled from mocked waitReconnect, got %v", err)
	}
	if waitCalls != 1 {
		t.Fatalf("expected waitReconnect to be called once, got %d", waitCalls)
	}
	if reconnectDelay != 5*time.Second {
		t.Fatalf("expected reconnect delay to be fixed 5s, got %s", reconnectDelay)
	}
}

func TestWithTokenUsesSharedHeaderConstant(t *testing.T) {
	ctx := withToken(context.Background(), "token-x")
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatalf("expected outgoing metadata")
	}

	tokens := md.Get(registryauth.HeaderSharedToken)
	if len(tokens) != 1 || tokens[0] != "token-x" {
		t.Fatalf("expected token in %q, got %#v", registryauth.HeaderSharedToken, tokens)
	}
}

func testConfig() config.Config {
	return config.Config{
		ConsoleGRPCTarget: "127.0.0.1:65535",
		SharedToken:       "test-token",
		HeartbeatInterval: 100 * time.Millisecond,
		CallTimeout:       10 * time.Millisecond,
		NodeName:          "node-test",
		ExecutorKind:      "docker",
		Version:           "test",
	}
}
