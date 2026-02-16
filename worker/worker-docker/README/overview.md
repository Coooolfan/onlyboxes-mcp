# Worker Docker Overview

`worker-docker` registers to console over gRPC and sends periodic heartbeats.

Defaults:
- Console target: `127.0.0.1:50051`
- Heartbeat interval: `5s`
