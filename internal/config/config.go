package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const FileName = "pb-agent.yaml"

type File struct {
	Version     int                   `yaml:"version" json:"version"`
	Connections map[string]Connection `yaml:"connections" json:"connections"`
}

type Connection struct {
	URL              string `yaml:"url" json:"url"`
	Environment      string `yaml:"environment" json:"environment"`
	Credential       string `yaml:"credential,omitempty" json:"credential,omitempty"`
	AllowInsecureTLS bool   `yaml:"allowInsecureTLS,omitempty" json:"allowInsecureTLS,omitempty"`
	MaxResponseItems int    `yaml:"maxResponseItems,omitempty" json:"maxResponseItems,omitempty"`
}

func Default() File {
	return File{Version: 1, Connections: map[string]Connection{}}
}

func Find(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, FileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func Load(path string) (File, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	var cfg File
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return File{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Version != 1 {
		return File{}, fmt.Errorf("unsupported config version %d", cfg.Version)
	}
	if cfg.Connections == nil {
		cfg.Connections = map[string]Connection{}
	}
	for name, connection := range cfg.Connections {
		environment, err := NormalizeEnvironment(connection.Environment)
		if err != nil {
			return File{}, fmt.Errorf("connection %q: %w", name, err)
		}
		connection.Environment = environment
		if err := ValidateConnection(name, connection); err != nil {
			return File{}, err
		}
		cfg.Connections[name] = connection
	}
	return cfg, nil
}

func Save(path string, cfg File) error {
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func NormalizeEnvironment(environment string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "dev", "development":
		return "development", nil
	case "test":
		return "test", nil
	case "stage", "staging":
		return "staging", nil
	case "prod", "production":
		return "production", nil
	default:
		return "", fmt.Errorf("environment must be dev, test, stage, or prod")
	}
}

func ValidateConnection(name string, connection Connection) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("connection name cannot be empty")
	}
	environment, err := NormalizeEnvironment(connection.Environment)
	if err != nil {
		return fmt.Errorf("connection %q: %w", name, err)
	}
	endpoint, err := url.Parse(connection.URL)
	if err != nil || endpoint.Hostname() == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") {
		return fmt.Errorf("connection %q URL must start with http:// or https://", name)
	}
	if endpoint.Scheme == "http" && !isLocalHost(endpoint.Hostname()) {
		return fmt.Errorf("connection %q uses public plaintext HTTP; use HTTPS", name)
	}
	if endpoint.Scheme == "http" && environment != "development" && environment != "test" {
		return fmt.Errorf("connection %q uses plaintext HTTP outside development/test", name)
	}
	if connection.AllowInsecureTLS && environment != "development" && environment != "test" {
		return fmt.Errorf("connection %q disables TLS verification outside development/test", name)
	}
	return nil
}

func isLocalHost(host string) bool {
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsPrivate())
}

func Resolve(path, name string) (Connection, string, error) {
	cfg, err := Load(path)
	if err != nil {
		return Connection{}, "", err
	}
	connection, ok := cfg.Connections[name]
	if !ok {
		return Connection{}, "", fmt.Errorf("connection %q is not configured", name)
	}
	return connection, name, nil
}
