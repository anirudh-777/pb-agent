package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anirudh-777/pb-agent/internal/config"
)

func TestCapabilitiesEnvelope(t *testing.T) {
	var stdout bytes.Buffer
	code := Run([]string{"capabilities"}, bytes.NewBuffer(nil), &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit code = %d, output = %s", code, stdout.String())
	}
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["schemaVersion"] != "1" || result["ok"] != true {
		t.Fatalf("unexpected envelope: %#v", result)
	}
}

func TestTokenHelpExplainsGenerationAndSafeStorage(t *testing.T) {
	var stdout bytes.Buffer
	code := Run([]string{"connection", "token-help"}, bytes.NewBuffer(nil), &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit code = %d, output = %s", code, stdout.String())
	}
	var result struct {
		Data struct {
			DashboardSteps []string `json:"dashboardSteps"`
			Documentation  string   `json:"documentation"`
			StoreCommand   string   `json:"storeCommand"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Data.DashboardSteps) < 4 {
		t.Fatalf("missing dashboard instructions: %#v", result.Data)
	}
	if result.Data.Documentation == "" || !strings.Contains(result.Data.StoreCommand, "--token-stdin") {
		t.Fatalf("missing safe token guidance: %#v", result.Data)
	}
}

func TestTokenHelpHumanOutput(t *testing.T) {
	var stdout bytes.Buffer
	code := Run([]string{"--human", "connection", "token-help"}, bytes.NewBuffer(nil), &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit code = %d, output = %s", code, stdout.String())
	}
	got := stdout.String()
	for _, expected := range []string{
		"Generate a PocketBase token for pb-agent",
		"1. Open the PocketBase Dashboard",
		"Store it securely",
		"Security",
		"Revocation",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("human output missing %q:\n%s", expected, got)
		}
	}
	if strings.Contains(got, `"schemaVersion"`) {
		t.Fatalf("human output contains JSON envelope:\n%s", got)
	}
}

func TestPlanApplyRecordUpdate(t *testing.T) {
	t.Setenv("PB_AGENT_STATE_DIR", t.TempDir())
	current := map[string]any{"id": "one", "name": "Before", "updated": "2026-07-23 00:00:00.000Z"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/collections/posts/records/one":
			_ = json.NewEncoder(w).Encode(current)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/collections/posts/records/one":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			current["name"] = body["name"]
			_ = json.NewEncoder(w).Encode(current)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, config.FileName)
	cfg := config.Default()
	cfg.Connections["default"] = config.Connection{URL: server.URL, Environment: "development"}
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(dir, "update.json")
	if err := os.WriteFile(dataPath, []byte(`{"name":"After"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var planned bytes.Buffer
	code := Run([]string{"--config", cfgPath, "plan", "record-update", "--collection", "posts", "--id", "one", "--data-file", dataPath}, bytes.NewBuffer(nil), &planned, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("plan exit = %d, output = %s", code, planned.String())
	}
	var envelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(planned.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	var applied bytes.Buffer
	code = Run([]string{"--config", cfgPath, "apply", "--plan", envelope.Data.ID}, bytes.NewBuffer(nil), &applied, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("apply exit = %d, output = %s", code, applied.String())
	}
	if current["name"] != "After" {
		t.Fatalf("record was not updated: %#v", current)
	}
}

func TestProductionPlanCannotApplyWithoutGrant(t *testing.T) {
	t.Setenv("PB_AGENT_STATE_DIR", t.TempDir())
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, config.FileName)
	cfg := config.Default()
	cfg.Connections["prod"] = config.Connection{URL: "https://example.invalid", Environment: "production"}
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(dir, "create.json")
	_ = os.WriteFile(dataPath, []byte(`{"name":"No"}`), 0o600)
	var planned bytes.Buffer
	code := Run([]string{"--config", cfgPath, "--connection", "prod", "plan", "record-create", "--collection", "posts", "--data-file", dataPath}, bytes.NewBuffer(nil), &planned, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("plan failed: %s", planned.String())
	}
	var envelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(planned.Bytes(), &envelope)
	var applied bytes.Buffer
	code = Run([]string{"--config", cfgPath, "--connection", "prod", "apply", "--plan", envelope.Data.ID}, bytes.NewBuffer(nil), &applied, &bytes.Buffer{})
	if code != 4 {
		t.Fatalf("apply exit = %d, output = %s", code, applied.String())
	}
}
