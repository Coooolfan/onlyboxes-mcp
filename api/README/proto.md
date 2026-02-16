# Proto Guide

## Files
- `proto/registry/v1/registry.proto`: shared worker registry API.
- `gen/go`: generated Go code.

## Prerequisites
Install generators:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`.

## Generate

```bash
./scripts/gen-go.sh
```

## Compatibility Rules
- Only append new fields with new tag numbers.
- Never delete or reuse field tags.
- Keep existing service/method names stable for backward compatibility.
