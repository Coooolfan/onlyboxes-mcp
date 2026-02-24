# Worker Sys Overview

`worker-sys` connects to console over gRPC bidi stream `Connect`, sends hello (`worker_secret`), sends periodic heartbeats, and handles `computerUse` command dispatch/result in the same stream.

Security warning (high risk):
- `computerUse` runs shell directly on the worker host (`/bin/sh -lc`).
- this worker is **not container-sandboxed**; commands can read/modify host files and processes under the worker OS account.
- run only on dedicated hosts with strict OS-level isolation and least-privilege service accounts.
- do not deploy on shared machines.
- console gRPC has no built-in TLS/mTLS; plaintext transport can expose `worker_secret`.
- place console gRPC behind trusted private networking or encrypted tunnels.

Required identity:
- `WORKER_ID`
- `WORKER_SECRET`

These values are returned by `console` when calling `POST /api/v1/workers`.
`WORKER_SECRET` is returned once at creation time; if lost, delete and recreate the worker.

Worker type and capability contract:
- worker type is `worker-sys`.
- hello declares only one capability: `computerUse`.
- `computerUse.max_inflight` is fixed to `1`.
- console enforces that `worker-sys` cannot register any other capability.

`computerUse` behavior:
- expected payload: `{"command":"..."}`
- `command` is required and executed via `/bin/sh -lc`.
- output fields:
  - `stdout`
  - `stderr`
  - `exit_code`
  - `stdout_truncated`
  - `stderr_truncated`
- non-zero process exit is returned in `exit_code` (not a command error by itself).
- output truncation is per stream and controlled by `WORKER_COMPUTER_USE_OUTPUT_LIMIT_BYTES`.

Defaults:
- Console target: `127.0.0.1:50051`
- Heartbeat interval: `5s`
- Heartbeat jitter: `20%`
- Output limit: `1048576` bytes per stream (`stdout`/`stderr`)

Config env:
- `WORKER_CONSOLE_GRPC_TARGET`
- `WORKER_CONSOLE_INSECURE`
- `WORKER_ID`
- `WORKER_SECRET`
- `WORKER_NODE_NAME`
- `WORKER_VERSION`
- `WORKER_LABELS`
- `WORKER_HEARTBEAT_INTERVAL_SEC`
- `WORKER_HEARTBEAT_JITTER_PCT`
- `WORKER_CALL_TIMEOUT_SEC`
- `WORKER_COMPUTER_USE_OUTPUT_LIMIT_BYTES`
