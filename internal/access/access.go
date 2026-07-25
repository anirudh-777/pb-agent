package access

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/anirudh-777/pb-agent/internal/policy"
	"github.com/anirudh-777/pb-agent/internal/state"
)

func path(connection string) (string, error) {
	if err := state.SafeID(connection); err != nil {
		return "", err
	}
	dir, err := state.Dir()
	if err != nil {
		return "", err
	}
	grants := filepath.Join(dir, "grants")
	if err := os.MkdirAll(grants, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(grants, connection+".json"), nil
}

func Save(grant policy.Grant) error {
	target, err := path(grant.Connection)
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(grant, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(target, raw, 0o600)
}

func Load(connection string) (*policy.Grant, error) {
	target, err := path(connection)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var grant policy.Grant
	if err := json.Unmarshal(raw, &grant); err != nil {
		return nil, err
	}
	if !time.Now().Before(grant.ExpiresAt) {
		_ = os.Remove(target)
		return nil, nil
	}
	return &grant, nil
}

func Revoke(connection string) error {
	target, err := path(connection)
	if err != nil {
		return err
	}
	err = os.Remove(target)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
