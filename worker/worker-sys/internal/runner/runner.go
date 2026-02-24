package runner

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/onlyboxes/onlyboxes/worker/worker-sys/internal/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	minHeartbeatInterval             = 1 * time.Second
	initialReconnectDelay            = 1 * time.Second
	maxReconnectDelay                = 15 * time.Second
	computerUseCapabilityName        = "computeruse"
	computerUseCapabilityDeclared    = "computerUse"
	computerUseCapabilityMaxInflight = 1
)

var waitReconnect = waitReconnectDelay
var applyJitter = jitterDuration

func Run(ctx context.Context, cfg config.Config) error {
	if strings.TrimSpace(cfg.WorkerID) == "" {
		return errors.New("WORKER_ID is required")
	}
	if strings.TrimSpace(cfg.WorkerSecret) == "" {
		return errors.New("WORKER_SECRET is required")
	}

	executor := newComputerUseExecutor(computerUseExecutorConfig{
		OutputLimitBytes: cfg.ComputerUseOutputLimitByte,
	})
	originalRunComputerUse := runComputerUse
	runComputerUse = executor.Execute
	defer func() {
		runComputerUse = originalRunComputerUse
	}()

	reconnectDelay := initialReconnectDelay
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := runSession(ctx, cfg)
		if err == nil {
			return nil
		}

		if errCtx := ctx.Err(); errCtx != nil {
			return errCtx
		}

		if status.Code(err) == codes.FailedPrecondition {
			log.Printf("registry session replaced for node_id=%s, reconnecting", cfg.WorkerID)
			reconnectDelay = initialReconnectDelay
		} else {
			log.Printf("registry session interrupted: %v", err)
		}

		if err := waitReconnect(ctx, reconnectDelay); err != nil {
			return err
		}
		reconnectDelay = nextReconnectDelay(reconnectDelay)
	}
}

func nextReconnectDelay(current time.Duration) time.Duration {
	if current <= 0 {
		return initialReconnectDelay
	}
	next := current * 2
	if next > maxReconnectDelay {
		return maxReconnectDelay
	}
	return next
}

func waitReconnectDelay(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		delay = initialReconnectDelay
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
