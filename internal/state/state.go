package state

import (
	"fmt"
	"os"
	"path/filepath"
)

func Dir() (string, error) {
	base := os.Getenv("PB_AGENT_STATE_DIR")
	if base == "" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(base, "pb-agent")
	}
	if err := os.MkdirAll(base, 0o700); err != nil {
		return "", err
	}
	return base, nil
}

func PlansDir() (string, error) {
	base, err := Dir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(base, "plans")
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", err
	}
	return path, nil
}

func SafeID(value string) error {
	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '_' && char != '-' {
			return fmt.Errorf("invalid identifier")
		}
	}
	if value == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	return nil
}
