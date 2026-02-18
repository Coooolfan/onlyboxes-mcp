package config

import (
	"reflect"
	"testing"
)

func TestLoadParsesMCPAllowedTokens(t *testing.T) {
	t.Setenv("CONSOLE_MCP_ALLOWED_TOKENS", " token-a,token-b , token-a, ,token-c,, ")

	cfg := Load()
	want := []string{"token-a", "token-b", "token-c"}
	if !reflect.DeepEqual(cfg.MCPAllowedTokens, want) {
		t.Fatalf("expected tokens=%v, got %v", want, cfg.MCPAllowedTokens)
	}
}

func TestLoadMCPAllowedTokensEmptyWhenUnset(t *testing.T) {
	t.Setenv("CONSOLE_MCP_ALLOWED_TOKENS", "")

	cfg := Load()
	if len(cfg.MCPAllowedTokens) != 0 {
		t.Fatalf("expected empty token list, got %v", cfg.MCPAllowedTokens)
	}
}

func TestLoadMCPAllowedTokensEmptyWhenWhitespaceOnly(t *testing.T) {
	t.Setenv("CONSOLE_MCP_ALLOWED_TOKENS", " , ,   , ")

	cfg := Load()
	if len(cfg.MCPAllowedTokens) != 0 {
		t.Fatalf("expected empty token list, got %v", cfg.MCPAllowedTokens)
	}
}
