package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"github.com/onlyboxes/onlyboxes/worker/worker-sys/internal/config"
	"github.com/onlyboxes/onlyboxes/worker/worker-sys/internal/runner"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runner.Run(ctx, cfg); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			log.Printf("worker stopped: %v", err)
			return
		}
		log.Fatalf("worker stopped with error: %v", err)
	}
}
