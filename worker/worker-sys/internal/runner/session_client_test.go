package runner

import (
	"context"
	"testing"
	"time"

	registryv1 "github.com/onlyboxes/onlyboxes/api/gen/go/registry/v1"
)

func TestCommandExecSlotAcquireRelease(t *testing.T) {
	commandExecSlots := make(chan struct{}, commandExecSlotCapacity)
	commandExecSlots <- struct{}{}

	if !tryAcquireCommandSlot(commandExecSlots) {
		t.Fatalf("expected first slot acquire to succeed")
	}
	if tryAcquireCommandSlot(commandExecSlots) {
		t.Fatalf("expected second slot acquire to fail while slot is held")
	}

	releaseCommandSlot(commandExecSlots)
	if !tryAcquireCommandSlot(commandExecSlots) {
		t.Fatalf("expected slot acquire to succeed after release")
	}
}

func TestHandleCommandDispatchBusyReturnsSessionBusyWithoutExecution(t *testing.T) {
	commandExecSlots := make(chan struct{}, commandExecSlotCapacity)
	commandExecSlots <- struct{}{}
	if !tryAcquireCommandSlot(commandExecSlots) {
		t.Fatalf("expected to acquire slot for busy-state setup")
	}
	defer releaseCommandSlot(commandExecSlots)

	outbound := make(chan *registryv1.ConnectRequest, 1)
	errCh := make(chan error, 1)
	executed := make(chan struct{}, 1)

	ok := handleCommandDispatch(
		context.Background(),
		outbound,
		errCh,
		commandExecSlots,
		&registryv1.CommandDispatch{
			CommandId:   "cmd-busy",
			Capability:  computerUseCapabilityDeclared,
			PayloadJson: []byte(`{"command":"pwd"}`),
		},
		func(context.Context, *registryv1.CommandDispatch) *registryv1.ConnectRequest {
			executed <- struct{}{}
			return commandErrorResult("cmd-busy", "execution_failed", "unexpected execution")
		},
	)
	if !ok {
		t.Fatalf("expected dispatch handling to continue")
	}

	select {
	case <-executed:
		t.Fatalf("execute function should not run while slot is busy")
	default:
	}

	select {
	case err := <-errCh:
		t.Fatalf("expected no session error, got %v", err)
	default:
	}

	select {
	case req := <-outbound:
		result := req.GetCommandResult()
		if result == nil {
			t.Fatalf("expected command_result payload")
		}
		if result.GetCommandId() != "cmd-busy" {
			t.Fatalf("expected command_id cmd-busy, got %q", result.GetCommandId())
		}
		if result.GetError() == nil {
			t.Fatalf("expected session_busy command error")
		}
		if result.GetError().GetCode() != sessionBusyErrorCode {
			t.Fatalf("expected error code %q, got %q", sessionBusyErrorCode, result.GetError().GetCode())
		}
		if result.GetError().GetMessage() != sessionBusyErrorMessage {
			t.Fatalf("expected error message %q, got %q", sessionBusyErrorMessage, result.GetError().GetMessage())
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for session_busy result")
	}
}

func TestHandleCommandDispatchRunsExecutionAndReleasesSlot(t *testing.T) {
	commandExecSlots := make(chan struct{}, commandExecSlotCapacity)
	commandExecSlots <- struct{}{}

	outbound := make(chan *registryv1.ConnectRequest, 1)
	errCh := make(chan error, 1)
	executed := make(chan struct{}, 1)

	ok := handleCommandDispatch(
		context.Background(),
		outbound,
		errCh,
		commandExecSlots,
		&registryv1.CommandDispatch{
			CommandId:   "cmd-run",
			Capability:  computerUseCapabilityDeclared,
			PayloadJson: []byte(`{"command":"pwd"}`),
		},
		func(context.Context, *registryv1.CommandDispatch) *registryv1.ConnectRequest {
			executed <- struct{}{}
			return commandErrorResult("cmd-run", "execution_failed", "forced test failure")
		},
	)
	if !ok {
		t.Fatalf("expected dispatch handling to continue")
	}

	select {
	case <-executed:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for execute function")
	}

	select {
	case req := <-outbound:
		result := req.GetCommandResult()
		if result == nil {
			t.Fatalf("expected command_result payload")
		}
		if result.GetCommandId() != "cmd-run" {
			t.Fatalf("expected command_id cmd-run, got %q", result.GetCommandId())
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for command result enqueue")
	}

	select {
	case err := <-errCh:
		t.Fatalf("expected no session error, got %v", err)
	default:
	}

	if !tryAcquireCommandSlot(commandExecSlots) {
		t.Fatalf("expected slot to be released after execution")
	}
}
