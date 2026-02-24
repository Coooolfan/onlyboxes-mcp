package config

import (
	"testing"

	"github.com/onlyboxes/onlyboxes/worker/worker-sys/internal/buildinfo"
)

func TestLoadUsesBuildVersionByDefault(t *testing.T) {
	t.Setenv("WORKER_VERSION", "")

	cfg := Load()
	if cfg.Version != buildinfo.Version {
		t.Fatalf("expected default worker version %q, got %q", buildinfo.Version, cfg.Version)
	}
}

func TestLoadSupportsCustomVersion(t *testing.T) {
	t.Setenv("WORKER_VERSION", "v1.2.3-custom")

	cfg := Load()
	if cfg.Version != "v1.2.3-custom" {
		t.Fatalf("expected custom worker version, got %q", cfg.Version)
	}
}

func TestLoadUsesComputerUseOutputLimitEnv(t *testing.T) {
	t.Setenv("WORKER_COMPUTER_USE_OUTPUT_LIMIT_BYTES", "2048")

	cfg := Load()
	if cfg.ComputerUseOutputLimitByte != 2048 {
		t.Fatalf("expected output limit 2048, got %d", cfg.ComputerUseOutputLimitByte)
	}
}
