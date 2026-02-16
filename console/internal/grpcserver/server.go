package grpcserver

import (
	registryv1 "github.com/onlyboxes/onlyboxes/api/gen/go/registry/v1"
	"google.golang.org/grpc"
)

func NewServer(sharedToken string, service registryv1.WorkerRegistryServiceServer) *grpc.Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryTokenAuthInterceptor(sharedToken)),
	)
	registryv1.RegisterWorkerRegistryServiceServer(server, service)
	return server
}
