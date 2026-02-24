package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	computerUseNotReadyMessage    = "computerUse executor is unavailable"
	computerUseCodeInvalidPayload = "invalid_payload"
)

type computerUsePayload struct {
	Command string `json:"command"`
}

type computerUseRequest struct {
	Command string
}

type computerUseRunResult struct {
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExitCode        int    `json:"exit_code"`
	StdoutTruncated bool   `json:"stdout_truncated"`
	StderrTruncated bool   `json:"stderr_truncated"`
}

type computerUseError struct {
	code    string
	message string
}

func (e *computerUseError) Error() string {
	if e == nil {
		return "computerUse execution failed"
	}
	return e.message
}

func (e *computerUseError) Code() string {
	if e == nil {
		return ""
	}
	return e.code
}

func newComputerUseError(code string, message string) *computerUseError {
	return &computerUseError{
		code:    strings.TrimSpace(code),
		message: strings.TrimSpace(message),
	}
}

type computerUseExecutorConfig struct {
	OutputLimitBytes int
}

type computerUseExecutor struct {
	outputLimitBytes int
}

func newComputerUseExecutor(cfg computerUseExecutorConfig) *computerUseExecutor {
	outputLimit := cfg.OutputLimitBytes
	if outputLimit <= 0 {
		outputLimit = 1024 * 1024
	}

	return &computerUseExecutor{
		outputLimitBytes: outputLimit,
	}
}

func (e *computerUseExecutor) Execute(ctx context.Context, req computerUseRequest) (computerUseRunResult, error) {
	if e == nil {
		return computerUseRunResult{}, newComputerUseError("execution_failed", computerUseNotReadyMessage)
	}

	command := strings.TrimSpace(req.Command)
	if command == "" {
		return computerUseRunResult{}, newComputerUseError(computerUseCodeInvalidPayload, "command is required")
	}

	execCmd := exec.CommandContext(ctx, "/bin/sh", "-lc", command)
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	execCmd.Stdout = &stdoutBuf
	execCmd.Stderr = &stderrBuf

	err := execCmd.Run()
	exitCode := 0
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return computerUseRunResult{}, err
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return computerUseRunResult{}, fmt.Errorf("shell execution failed: %w", err)
		}
	} else if execCmd.ProcessState != nil {
		exitCode = execCmd.ProcessState.ExitCode()
	}

	stdout, stdoutTruncated := truncateByBytes(stdoutBuf.String(), e.outputLimitBytes)
	stderr, stderrTruncated := truncateByBytes(stderrBuf.String(), e.outputLimitBytes)

	return computerUseRunResult{
		Stdout:          stdout,
		Stderr:          stderr,
		ExitCode:        exitCode,
		StdoutTruncated: stdoutTruncated,
		StderrTruncated: stderrTruncated,
	}, nil
}

func truncateByBytes(value string, maxBytes int) (string, bool) {
	if maxBytes <= 0 {
		return value, false
	}
	if len(value) <= maxBytes {
		return value, false
	}
	return value[:maxBytes], true
}
