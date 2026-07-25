package config

import (
	"errors"
	"fmt"
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
		if err := ValidateConnection(name, connection); err != nil {
			return File{}, err
		}
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

func ValidateConnection(name string, connection Connection) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("connection name cannot be empty")
	}
	switch connection.Environment {
	case "development", "test", "staging", "production":
	default:
		return fmt.Errorf("connection %q has invalid environment %q", name, connection.Environment)
	}
	if !strings.HasPrefix(connection.URL, "http://") && !strings.HasPrefix(connection.URL, "https://") {
		return fmt.Errorf("connection %q URL must start with http:// or https://", name)
	}
	if strings.HasPrefix(connection.URL, "http://") && connection.Environment != "development" && connection.Environment != "test" {
		return fmt.Errorf("connection %q uses plaintext HTTP outside development/test", name)
	}
	if connection.AllowInsecureTLS && connection.Environment != "development" && connection.Environment != "test" {
		return fmt.Errorf("connection %q disables TLS verification outside development/test", name)
	}
	return nil
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
