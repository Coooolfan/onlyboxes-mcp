package grpcserver

import (
	"context"
	"strings"

	"github.com/onlyboxes/onlyboxes/api/pkg/registryauth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const HeaderSharedToken = registryauth.HeaderSharedToken

func UnaryTokenAuthInterceptor(sharedToken string) grpc.UnaryServerInterceptor {
	trimmedConfig := strings.TrimSpace(sharedToken)
	allowedTokens := parseAllowedTokens(trimmedConfig)
	authDisabled := trimmedConfig == ""

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if authDisabled {
			return handler(ctx, req)
		}
		if len(allowedTokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "auth token configuration is invalid")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		tokens := md.Get(HeaderSharedToken)
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing auth token")
		}

		for _, token := range tokens {
			if _, ok := allowedTokens[token]; ok {
				return handler(ctx, req)
			}
		}
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}
}

func parseAllowedTokens(raw string) map[string]struct{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]struct{}{}
	}

	allowed := make(map[string]struct{})
	for _, token := range strings.Split(raw, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		allowed[token] = struct{}{}
	}
	return allowed
}
