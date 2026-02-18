package runner

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	terminalResourceCapabilityName     = "terminalresource"
	terminalResourceCapabilityDeclared = "terminalResource"
	terminalResourceActionValidate     = "validate"
	terminalResourceActionRead         = "read"
	terminalResourceCodeFileNotFound   = "file_not_found"
	terminalResourceCodePathIsDir      = "path_is_directory"
	terminalResourceCodeFileTooLarge   = "file_too_large"
)

type terminalResourcePayload struct {
	SessionID string `json:"session_id"`
	FilePath  string `json:"file_path"`
	Action    string `json:"action,omitempty"`
}

type terminalResourceRequest struct {
	SessionID string
	FilePath  string
	Action    string
}

type terminalResourceRunResult struct {
	SessionID string `json:"session_id"`
	FilePath  string `json:"file_path"`
	MIMEType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	Blob      []byte `json:"blob,omitempty"`
}

type terminalResourceProbeResult struct {
	Error    string `json:"error,omitempty"`
	Message  string `json:"message,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
	Size     int64  `json:"size_bytes"`
	Blob     string `json:"blob,omitempty"`
}

const terminalResourceProbeScript = `
import argparse
import base64
import json
import mimetypes
import os
import sys

parser = argparse.ArgumentParser()
parser.add_argument("--action", choices=["validate", "read"], default="validate")
parser.add_argument("--file-path", required=True)
parser.add_argument("--max-read-bytes", type=int, required=True)
args = parser.parse_args()

target = args.file_path
if not os.path.exists(target):
    print(json.dumps({"error": "file_not_found", "message": "file not found"}))
    sys.exit(10)
if os.path.isdir(target):
    print(json.dumps({"error": "path_is_directory", "message": "path is directory"}))
    sys.exit(11)

size_bytes = os.path.getsize(target)
mime_type, _ = mimetypes.guess_type(target)
if not mime_type:
    mime_type = "application/octet-stream"

if args.action == "validate":
    print(json.dumps({"mime_type": mime_type, "size_bytes": size_bytes}))
    sys.exit(0)

limit = args.max_read_bytes
if size_bytes > limit:
    print(json.dumps({
        "error": "file_too_large",
        "message": "file exceeds read limit",
        "mime_type": mime_type,
        "size_bytes": size_bytes,
    }))
    sys.exit(12)

with open(target, "rb") as fh:
    content = fh.read(limit + 1)
if len(content) > limit:
    print(json.dumps({
        "error": "file_too_large",
        "message": "file exceeds read limit",
        "mime_type": mime_type,
        "size_bytes": len(content),
    }))
    sys.exit(12)

print(json.dumps({
    "mime_type": mime_type,
    "size_bytes": len(content),
    "blob": base64.b64encode(content).decode("ascii"),
}))
`

func (m *terminalSessionManager) ResolveResource(ctx context.Context, req terminalResourceRequest) (terminalResourceRunResult, error) {
	if m == nil {
		return terminalResourceRunResult{}, newTerminalExecError("execution_failed", terminalExecNotReadyMessage)
	}

	sessionID := strings.TrimSpace(req.SessionID)
	filePath := strings.TrimSpace(req.FilePath)
	if sessionID == "" || filePath == "" {
		return terminalResourceRunResult{}, newTerminalExecError(terminalExecCodeInvalidPayload, "session_id and file_path are required")
	}

	action := normalizeTerminalResourceAction(req.Action)
	if action == "" {
		return terminalResourceRunResult{}, newTerminalExecError(terminalExecCodeInvalidPayload, "action must be validate or read")
	}

	m.mu.Lock()
	session, ok := m.sessions[sessionID]
	if !ok || session == nil {
		m.mu.Unlock()
		return terminalResourceRunResult{}, newTerminalExecError(terminalExecCodeSessionNotFound, terminalExecNoSessionMessage)
	}
	if session.busy {
		m.mu.Unlock()
		return terminalResourceRunResult{}, newTerminalExecError(terminalExecCodeSessionBusy, terminalExecBusyMessage)
	}
	session.busy = true
	containerName := session.containerName
	m.mu.Unlock()

	execResult := runDockerCommand(ctx, terminalExecDockerResourceArgs(containerName, action, filePath, m.outputLimitBytes)...)
	if execResult.Err != nil {
		if errors.Is(execResult.Err, context.DeadlineExceeded) || errors.Is(execResult.Err, context.Canceled) {
			m.destroySession(sessionID)
			return terminalResourceRunResult{}, execResult.Err
		}
		m.markSessionIdle(sessionID)
		return terminalResourceRunResult{}, fmt.Errorf("docker exec failed: %w", execResult.Err)
	}
	if isNoSuchContainerMessage(execResult.Stderr) {
		m.destroySession(sessionID)
		return terminalResourceRunResult{}, newTerminalExecError(terminalExecCodeSessionNotFound, terminalExecNoSessionMessage)
	}
	if _, ok := m.markSessionIdle(sessionID); !ok {
		return terminalResourceRunResult{}, newTerminalExecError(terminalExecCodeSessionNotFound, terminalExecNoSessionMessage)
	}

	probe, err := decodeTerminalResourceProbeOutput(execResult.Stdout)
	if err != nil {
		return terminalResourceRunResult{}, fmt.Errorf("invalid terminalResource result: %w", err)
	}

	if code := strings.TrimSpace(probe.Error); code != "" {
		message := terminalResourceErrorMessage(code, probe.Message)
		return terminalResourceRunResult{}, newTerminalExecError(code, message)
	}
	if execResult.ExitCode != 0 {
		return terminalResourceRunResult{}, fmt.Errorf(
			"docker exec failed: %s",
			dockerCommandFailureMessage("exit code", execResult.ExitCode, execResult.Stderr),
		)
	}

	mimeType := strings.TrimSpace(probe.MIMEType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	result := terminalResourceRunResult{
		SessionID: sessionID,
		FilePath:  filePath,
		MIMEType:  mimeType,
		SizeBytes: probe.Size,
	}
	if action != terminalResourceActionRead {
		return result, nil
	}

	blobValue := strings.TrimSpace(probe.Blob)
	if blobValue == "" {
		result.Blob = []byte{}
		return result, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(blobValue)
	if err != nil {
		return terminalResourceRunResult{}, fmt.Errorf("decode resource blob: %w", err)
	}
	result.Blob = decoded
	return result, nil
}

func terminalExecDockerResourceArgs(containerName string, action string, filePath string, maxReadBytes int) []string {
	limit := maxReadBytes
	if limit <= 0 {
		limit = 1
	}
	return []string{
		"exec",
		containerName,
		"python",
		"-c",
		terminalResourceProbeScript,
		"--action",
		action,
		"--file-path",
		filePath,
		"--max-read-bytes",
		strconv.Itoa(limit),
	}
}

func normalizeTerminalResourceAction(action string) string {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "":
		return terminalResourceActionValidate
	case terminalResourceActionValidate:
		return terminalResourceActionValidate
	case terminalResourceActionRead:
		return terminalResourceActionRead
	default:
		return ""
	}
}

func decodeTerminalResourceProbeOutput(stdout string) (terminalResourceProbeResult, error) {
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" {
		return terminalResourceProbeResult{}, errors.New("empty output")
	}
	decoded := terminalResourceProbeResult{}
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return terminalResourceProbeResult{}, err
	}
	return decoded, nil
}

func terminalResourceErrorMessage(code string, fallback string) string {
	if trimmed := strings.TrimSpace(fallback); trimmed != "" {
		return trimmed
	}
	switch strings.TrimSpace(code) {
	case terminalResourceCodeFileNotFound:
		return "file not found"
	case terminalResourceCodePathIsDir:
		return "path is directory"
	case terminalResourceCodeFileTooLarge:
		return "file exceeds read limit"
	default:
		return "terminal resource operation failed"
	}
}
