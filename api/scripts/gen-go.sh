#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="$ROOT_DIR/proto"
MODULE="github.com/onlyboxes/onlyboxes/api"

mkdir -p "$ROOT_DIR/gen/go"

protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$ROOT_DIR" \
  --go_opt="module=$MODULE" \
  --go-grpc_out="$ROOT_DIR" \
  --go-grpc_opt="module=$MODULE" \
  "$PROTO_DIR/registry/v1/registry.proto"

echo "Generated Go protobuf files under $ROOT_DIR/gen/go"
