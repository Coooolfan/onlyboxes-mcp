package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPAddr             = ":8089"
	defaultGRPCAddr             = ":50051"
	defaultSharedToken          = "onlyboxes-dev-token"
	defaultOfflineTTLSec        = 15
	defaultHeartbeatIntervalSec = 5
)

type Config struct {
	HTTPAddr                string
	GRPCAddr                string
	GRPCSharedToken         string
	UsingDefaultSharedToken bool
	OfflineTTL              time.Duration
	HeartbeatIntervalSec    int32
}

func Load() Config {
	offlineTTLSec := parsePositiveIntEnv("CONSOLE_OFFLINE_TTL_SEC", defaultOfflineTTLSec)
	sharedToken := getEnv("CONSOLE_GRPC_SHARED_TOKEN", defaultSharedToken)

	return Config{
		HTTPAddr:                getEnv("CONSOLE_HTTP_ADDR", defaultHTTPAddr),
		GRPCAddr:                getEnv("CONSOLE_GRPC_ADDR", defaultGRPCAddr),
		GRPCSharedToken:         sharedToken,
		UsingDefaultSharedToken: sharedToken == defaultSharedToken,
		OfflineTTL:              time.Duration(offlineTTLSec) * time.Second,
		HeartbeatIntervalSec:    defaultHeartbeatIntervalSec,
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func parsePositiveIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
