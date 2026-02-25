package config

import (
	"reflect"
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

func TestLoadUsesDefaultComputerUseWhitelistMode(t *testing.T) {
	t.Setenv("WORKER_COMPUTER_USE_COMMAND_WHITELIST_MODE", "")

	cfg := Load()
	if cfg.ComputerUseWhitelistMode != "exact" {
		t.Fatalf("expected default whitelist mode exact, got %q", cfg.ComputerUseWhitelistMode)
	}
}

func TestLoadNormalizesComputerUseWhitelistMode(t *testing.T) {
	t.Setenv("WORKER_COMPUTER_USE_COMMAND_WHITELIST_MODE", "  PREFIX ")

	cfg := Load()
	if cfg.ComputerUseWhitelistMode != "prefix" {
		t.Fatalf("expected whitelist mode prefix, got %q", cfg.ComputerUseWhitelistMode)
	}
}

func TestLoadFallsBackComputerUseWhitelistModeOnInvalidValue(t *testing.T) {
	t.Setenv("WORKER_COMPUTER_USE_COMMAND_WHITELIST_MODE", "bad_mode")

	cfg := Load()
	if cfg.ComputerUseWhitelistMode != "exact" {
		t.Fatalf("expected fallback whitelist mode exact, got %q", cfg.ComputerUseWhitelistMode)
	}
}

func TestLoadParsesComputerUseWhitelist(t *testing.T) {
	t.Setenv("WORKER_COMPUTER_USE_COMMAND_WHITELIST", `[" echo ","time","","echo"]`)

	cfg := Load()
	want := []string{"echo", "time"}
	if !reflect.DeepEqual(cfg.ComputerUseWhitelist, want) {
		t.Fatalf("unexpected whitelist: want=%v got=%v", want, cfg.ComputerUseWhitelist)
	}
}

func TestLoadUsesEmptyComputerUseWhitelistWhenInvalid(t *testing.T) {
	t.Setenv("WORKER_COMPUTER_USE_COMMAND_WHITELIST", `{"not":"array"}`)

	cfg := Load()
	if len(cfg.ComputerUseWhitelist) != 0 {
		t.Fatalf("expected empty whitelist, got %v", cfg.ComputerUseWhitelist)
	}
}
