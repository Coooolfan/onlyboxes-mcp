package runner

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	registryv1 "github.com/onlyboxes/onlyboxes/api/gen/go/registry/v1"
	"github.com/onlyboxes/onlyboxes/api/pkg/registryauth"
	"github.com/onlyboxes/onlyboxes/worker/worker-docker/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const heartbeatLogEvery = 12

var reconnectDelay = 5 * time.Second
var waitReconnect = waitReconnectDelay

func Run(ctx context.Context, cfg config.Config) error {
	nodeID := uuid.NewString()
	nodeName := strings.TrimSpace(cfg.NodeName)
	if nodeName == "" {
		nodeName = fmt.Sprintf("worker-docker-%s", nodeID[:8])
	}

	registerReq := &registryv1.RegisterRequest{
		NodeId:       nodeID,
		NodeName:     nodeName,
		ExecutorKind: cfg.ExecutorKind,
		Languages:    cfg.Languages,
		Labels:       cfg.Labels,
		Version:      cfg.Version,
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := runSession(ctx, cfg, registerReq, nodeID, nodeName)
		if err == nil {
			return nil
		}

		if errCtx := ctx.Err(); errCtx != nil {
			return errCtx
		}
		log.Printf("registry session interrupted: %v", err)
		if err := waitReconnect(ctx); err != nil {
			return err
		}
	}
}

func runSession(ctx context.Context, cfg config.Config, registerReq *registryv1.RegisterRequest, nodeID string, nodeName string) error {
	conn, err := dial(ctx, cfg)
	if err != nil {
		return fmt.Errorf("dial console: %w", err)
	}
	defer conn.Close()

	client := registryv1.NewWorkerRegistryServiceClient(conn)
	if err := register(ctx, client, cfg, registerReq); err != nil {
		return fmt.Errorf("register node %s: %w", nodeID, err)
	}

	log.Printf("worker registered: node_id=%s node_name=%s", nodeID, nodeName)
	if err := heartbeatLoop(ctx, client, cfg, nodeID); err != nil {
		return fmt.Errorf("heartbeat loop: %w", err)
	}
	return nil
}

func dial(ctx context.Context, cfg config.Config) (*grpc.ClientConn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return grpc.NewClient(
		cfg.ConsoleGRPCTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func register(ctx context.Context, client registryv1.WorkerRegistryServiceClient, cfg config.Config, req *registryv1.RegisterRequest) error {
	callCtx, cancel := context.WithTimeout(ctx, cfg.CallTimeout)
	defer cancel()
	_, err := client.Register(withToken(callCtx, cfg.SharedToken), req)
	return err
}

func heartbeatLoop(ctx context.Context, client registryv1.WorkerRegistryServiceClient, cfg config.Config, nodeID string) error {
	ticker := time.NewTicker(cfg.HeartbeatInterval)
	defer ticker.Stop()
	successCount := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			callCtx, cancel := context.WithTimeout(ctx, cfg.CallTimeout)
			_, err := client.Heartbeat(withToken(callCtx, cfg.SharedToken), &registryv1.HeartbeatRequest{NodeId: nodeID})
			cancel()
			if err != nil {
				log.Printf("heartbeat failed: node_id=%s err=%v", nodeID, err)
				return err
			}
			successCount++
			if successCount == 1 || successCount%heartbeatLogEvery == 0 {
				log.Printf("heartbeat ok: node_id=%s count=%d", nodeID, successCount)
			}
		}
	}
}

func withToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, registryauth.HeaderSharedToken, token)
}

func waitReconnectDelay(ctx context.Context) error {
	timer := time.NewTimer(reconnectDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
