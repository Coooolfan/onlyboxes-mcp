package runner

import (
	"context"
	"errors"
	"testing"

	registryv1 "github.com/onlyboxes/onlyboxes/api/gen/go/registry/v1"
)

func TestComputerUseExecutorExactModeAllowsExactMatch(t *testing.T) {
	executor := newComputerUseExecutor(computerUseExecutorConfig{
		OutputLimitBytes: 1024,
		WhitelistMode:    computerUseWhitelistModeExact,
		Whitelist:        []string{"echo"},
	})

	result, err := executor.Execute(context.Background(), computerUseRequest{Command: "echo"})
	if err != nil {
		t.Fatalf("expected command to be allowed, got error %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestComputerUseExecutorExactModeRejectsNonExactCommand(t *testing.T) {
	executor := newComputerUseExecutor(computerUseExecutorConfig{
		OutputLimitBytes: 1024,
		WhitelistMode:    computerUseWhitelistModeExact,
		Whitelist:        []string{"echo"},
	})

	_, err := executor.Execute(context.Background(), computerUseRequest{Command: "echo hi"})
	assertComputerUseErrorCode(t, err, computerUseCodeCommandNotAllowed)
}

func TestComputerUseExecutorPrefixModeAllowsCommand(t *testing.T) {
	executor := newComputerUseExecutor(computerUseExecutorConfig{
		OutputLimitBytes: 1024,
		WhitelistMode:    computerUseWhitelistModePrefix,
		Whitelist:        []string{"echo"},
	})

	result, err := executor.Execute(context.Background(), computerUseRequest{Command: "echo hi"})
	if err != nil {
		t.Fatalf("expected command to be allowed, got error %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestComputerUseExecutorPrefixModeRejectsNonMatchingCommand(t *testing.T) {
	executor := newComputerUseExecutor(computerUseExecutorConfig{
		OutputLimitBytes: 1024,
		WhitelistMode:    computerUseWhitelistModePrefix,
		Whitelist:        []string{"echo"},
	})

	_, err := executor.Execute(context.Background(), computerUseRequest{Command: "pwd"})
	assertComputerUseErrorCode(t, err, computerUseCodeCommandNotAllowed)
}

func TestComputerUseExecutorAllowAllModeIgnoresWhitelist(t *testing.T) {
	executor := newComputerUseExecutor(computerUseExecutorConfig{
		OutputLimitBytes: 1024,
		WhitelistMode:    computerUseWhitelistModeAllowAll,
		Whitelist:        []string{},
	})

	result, err := executor.Execute(context.Background(), computerUseRequest{Command: "printf test"})
	if err != nil {
		t.Fatalf("expected command to be allowed, got error %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestComputerUseExecutorEmptyWhitelistRejectsAllInExactMode(t *testing.T) {
	executor := newComputerUseExecutor(computerUseExecutorConfig{
		OutputLimitBytes: 1024,
		WhitelistMode:    computerUseWhitelistModeExact,
		Whitelist:        []string{},
	})

	_, err := executor.Execute(context.Background(), computerUseRequest{Command: "echo"})
	assertComputerUseErrorCode(t, err, computerUseCodeCommandNotAllowed)
}

func TestBuildCommandResultMapsCommandNotAllowedError(t *testing.T) {
	originalRunComputerUse := runComputerUse
	t.Cleanup(func() {
		runComputerUse = originalRunComputerUse
	})

	runComputerUse = func(_ context.Context, _ computerUseRequest) (computerUseRunResult, error) {
		return computerUseRunResult{}, newComputerUseError(computerUseCodeCommandNotAllowed, "command is blocked by whitelist policy")
	}

	req := buildCommandResult(&registryv1.CommandDispatch{
		CommandId:   "cmd-whitelist",
		Capability:  computerUseCapabilityDeclared,
		PayloadJson: []byte(`{"command":"echo"}`),
	})
	result := req.GetCommandResult()
	if result == nil {
		t.Fatalf("expected command_result payload")
	}
	if result.GetError() == nil {
		t.Fatalf("expected command error")
	}
	if result.GetError().GetCode() != computerUseCodeCommandNotAllowed {
		t.Fatalf("expected error code %q, got %q", computerUseCodeCommandNotAllowed, result.GetError().GetCode())
	}
}

func assertComputerUseErrorCode(t *testing.T, err error, wantCode string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error with code %q, got nil", wantCode)
	}

	var cuErr *computerUseError
	if !errors.As(err, &cuErr) {
		t.Fatalf("expected computerUseError, got %T", err)
	}
	if cuErr.Code() != wantCode {
		t.Fatalf("expected error code %q, got %q", wantCode, cuErr.Code())
	}
}
