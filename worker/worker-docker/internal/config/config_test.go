package config

import (
	"testing"

	"github.com/onlyboxes/onlyboxes/worker/worker-docker/internal/buildinfo"
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

func TestLoadUsesDefaultDockerImages(t *testing.T) {
	t.Setenv("WORKER_PYTHON_EXEC_DOCKER_IMAGE", "")
	t.Setenv("WORKER_TERMINAL_EXEC_DOCKER_IMAGE", "")

	cfg := Load()
	if cfg.PythonExecDockerImage != defaultPythonExecImage {
		t.Fatalf("expected default pythonExec image %q, got %q", defaultPythonExecImage, cfg.PythonExecDockerImage)
	}
	if cfg.TerminalExecDockerImage != defaultTerminalExecImage {
		t.Fatalf("expected default terminalExec image %q, got %q", defaultTerminalExecImage, cfg.TerminalExecDockerImage)
	}
}

func TestLoadSupportsCustomDockerImages(t *testing.T) {
	t.Setenv("WORKER_PYTHON_EXEC_DOCKER_IMAGE", "python:3.12-alpine")
	t.Setenv("WORKER_TERMINAL_EXEC_DOCKER_IMAGE", "debian:bookworm-slim")

	cfg := Load()
	if cfg.PythonExecDockerImage != "python:3.12-alpine" {
		t.Fatalf("expected custom pythonExec image, got %q", cfg.PythonExecDockerImage)
	}
	if cfg.TerminalExecDockerImage != "debian:bookworm-slim" {
		t.Fatalf("expected custom terminalExec image, got %q", cfg.TerminalExecDockerImage)
	}
}
