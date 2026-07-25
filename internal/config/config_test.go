package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeEnvironmentAliases(t *testing.T) {
	for input, want := range map[string]string{
		"dev":         "development",
		"development": "development",
		"stage":       "staging",
		"staging":     "staging",
		"prod":        "production",
		"production":  "production",
		"test":        "test",
	} {
		got, err := NormalizeEnvironment(input)
		if err != nil {
			t.Fatalf("NormalizeEnvironment(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("NormalizeEnvironment(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestValidateConnectionRejectsPublicPlaintextHTTP(t *testing.T) {
	connection := Connection{URL: "http://147.79.71.250", Environment: "development"}
	if err := ValidateConnection("default", connection); err == nil {
		t.Fatal("expected public plaintext HTTP to be rejected")
	}
}

func TestValidateConnectionAllowsPrivateDevelopmentHTTP(t *testing.T) {
	for _, endpoint := range []string{
		"http://127.0.0.1:8090",
		"http://localhost:8090",
		"http://192.168.1.10:8090",
	} {
		connection := Connection{URL: endpoint, Environment: "development"}
		if err := ValidateConnection("default", connection); err != nil {
			t.Fatalf("ValidateConnection(%q): %v", endpoint, err)
		}
	}
}

func TestLoadCanonicalizesEnvironmentAlias(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	raw := []byte("version: 1\nconnections:\n  default:\n    url: http://127.0.0.1:8090\n    environment: dev\n")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Connections["default"].Environment; got != "development" {
		t.Fatalf("environment = %q, want development", got)
	}
}
