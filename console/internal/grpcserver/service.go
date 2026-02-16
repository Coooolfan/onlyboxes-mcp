package grpcserver

import (
	"context"
	"strings"
	"time"

	registryv1 "github.com/onlyboxes/onlyboxes/api/gen/go/registry/v1"
	"github.com/onlyboxes/onlyboxes/console/internal/registry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RegistryService struct {
	registryv1.UnimplementedWorkerRegistryServiceServer

	store                *registry.Store
	heartbeatIntervalSec int32
	offlineTTLSec        int32
	nowFn                func() time.Time
}

const maxNodeIDLength = 128

func NewRegistryService(store *registry.Store, heartbeatIntervalSec int32, offlineTTLSec int32) *RegistryService {
	return &RegistryService{
		store:                store,
		heartbeatIntervalSec: heartbeatIntervalSec,
		offlineTTLSec:        offlineTTLSec,
		nowFn:                time.Now,
	}
}

func (s *RegistryService) Register(ctx context.Context, req *registryv1.RegisterRequest) (*registryv1.RegisterResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if err := validateNodeID(req.GetNodeId()); err != nil {
		return nil, err
	}

	now := s.nowFn()
	s.store.Upsert(req, now)
	return s.response(now), nil
}

func (s *RegistryService) Heartbeat(ctx context.Context, req *registryv1.HeartbeatRequest) (*registryv1.HeartbeatResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if err := validateNodeID(req.GetNodeId()); err != nil {
		return nil, err
	}

	now := s.nowFn()
	err := s.store.Touch(req.GetNodeId(), now)
	if err == registry.ErrNodeNotFound {
		return nil, status.Error(codes.NotFound, "node not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update heartbeat")
	}

	return &registryv1.HeartbeatResponse{
		ServerTimeUnixMs:     now.UnixMilli(),
		HeartbeatIntervalSec: s.heartbeatIntervalSec,
		OfflineTtlSec:        s.offlineTTLSec,
	}, nil
}

func (s *RegistryService) response(now time.Time) *registryv1.RegisterResponse {
	return &registryv1.RegisterResponse{
		ServerTimeUnixMs:     now.UnixMilli(),
		HeartbeatIntervalSec: s.heartbeatIntervalSec,
		OfflineTtlSec:        s.offlineTTLSec,
	}
}

func validateNodeID(nodeID string) error {
	if strings.TrimSpace(nodeID) == "" {
		return status.Error(codes.InvalidArgument, "node_id is required")
	}
	if len(nodeID) > maxNodeIDLength {
		return status.Error(codes.InvalidArgument, "node_id is too long")
	}
	return nil
}
