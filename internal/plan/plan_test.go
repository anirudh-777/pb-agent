package plan

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anirudh-777/pb-agent/internal/policy"
	"github.com/anirudh-777/pb-agent/internal/state"
)

func TestPlanRoundTripAndTamperDetection(t *testing.T) {
	t.Setenv("PB_AGENT_STATE_DIR", t.TempDir())
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	created, err := New("record.create", "dev", "fingerprint", "development", policy.Write, "records.write", "POST", "/api/records", []byte(`{"name":"Ada"}`), map[string]any{"name": "Ada"}, "", "", now)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := created.Payload()
	if err != nil || string(payload) != `{"name":"Ada"}` {
		t.Fatalf("payload = %q, err = %v", payload, err)
	}
	if created.EncryptedPayload == "" || created.EncryptedPayload == base64.RawStdEncoding.EncodeToString([]byte(`{"name":"Ada"}`)) {
		t.Fatal("stored payload was not encrypted")
	}
	created.RequestHash = "bad"
	if _, err := created.Payload(); err == nil {
		t.Fatal("expected tamper detection")
	}
}

func TestPlanValidation(t *testing.T) {
	t.Setenv("PB_AGENT_STATE_DIR", t.TempDir())
	now := time.Now().UTC()
	created, err := New("record.delete", "dev", "one", "development", policy.Destructive, "records.delete", "DELETE", "/api/records/1", nil, nil, "", "", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := created.Validate("two", "development", now); err == nil {
		t.Fatal("expected target mismatch")
	}
	if err := created.Validate("one", "production", now); err == nil {
		t.Fatal("expected environment mismatch")
	}
	if err := created.Validate("one", "development", now.Add(16*time.Minute)); err == nil {
		t.Fatal("expected expiry")
	}
}

func TestSavedPlanRejectsMetadataTampering(t *testing.T) {
	t.Setenv("PB_AGENT_STATE_DIR", t.TempDir())
	now := time.Now().UTC()
	created, err := New("record.create", "dev", "one", "development", policy.Write, "records.write", "POST", "/api/records", []byte(`{"name":"Ada"}`), nil, "", "", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(created); err != nil {
		t.Fatal(err)
	}
	dir, err := state.PlansDir()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, created.ID+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.Replace(raw, []byte(`"environment": "development"`), []byte(`"environment": "production"`), 1)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(created.ID); err == nil {
		t.Fatal("expected integrity failure")
	}
}

func TestAcquirePreventsConcurrentApply(t *testing.T) {
	t.Setenv("PB_AGENT_STATE_DIR", t.TempDir())
	release, err := Acquire("pln_test")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := Acquire("pln_test"); err == nil {
		t.Fatal("expected second acquisition to fail")
	}
}
