package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anirudh-777/pb-agent/internal/config"
	"github.com/anirudh-777/pb-agent/internal/output"
	"github.com/anirudh-777/pb-agent/internal/pocketbase"
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
	data, ok := result["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing capability data: %#v", result)
	}
	commands, ok := data["commands"].(map[string]any)
	if !ok || commands["records.list"] != "pb-agent records list --collection NAME --page PAGE --per-page N" {
		t.Fatalf("missing exact command grammar: %#v", data)
	}
}

func TestUnknownCommandIsReportedAsInvalidArguments(t *testing.T) {
	var stdout bytes.Buffer
	code := Run([]string{"collection", "list"}, bytes.NewBuffer(nil), &stdout, &bytes.Buffer{})
	if code != 2 {
		t.Fatalf("exit code = %d, output = %s", code, stdout.String())
	}
	var result struct {
		Command string `json:"command"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Command != "collection.list" || result.Error.Code != "invalid_arguments" {
		t.Fatalf("unexpected error envelope: %#v", result)
	}
}

func TestUnknownFlagIsReportedAgainstResolvedCommand(t *testing.T) {
	var stdout bytes.Buffer
	code := Run([]string{"records", "list", "--collection", "users", "--bogus"}, bytes.NewBuffer(nil), &stdout, &bytes.Buffer{})
	if code != 2 {
		t.Fatalf("exit code = %d, output = %s", code, stdout.String())
	}
	var result struct {
		Command string `json:"command"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Command != "records.list" || result.Error.Code != "invalid_arguments" {
		t.Fatalf("unexpected error envelope: %#v", result)
	}
}

func TestForbiddenPocketBaseResponseIsPolicyDenied(t *testing.T) {
	err := mapPBError(&pocketbase.APIError{Status: http.StatusForbidden, Message: "Batch requests are not allowed."})
	var cliErr *output.CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T", err)
	}
	if cliErr.Code != "policy_denied" || cliErr.ExitCode != 4 || cliErr.Message != "Batch requests are not allowed." {
		t.Fatalf("unexpected mapped error: %#v", cliErr)
	}
}

func TestRuntimeFailureRemainsInternalError(t *testing.T) {
	code := Run([]string{"capabilities"}, bytes.NewBuffer(nil), errorWriter{}, &bytes.Buffer{})
	if code != 1 {
		t.Fatalf("exit code = %d, want internal-error exit code 1", code)
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestTokenHelpExplainsGenerationAndSafeStorage(t *testing.T) {
	var stdout bytes.Buffer
	code := Run([]string{"connection", "token-help"}, bytes.NewBuffer(nil), &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit code = %d, output = %s", code, stdout.String())
	}
	var result struct {
		Data struct {
			DashboardSteps    []string `json:"dashboardSteps"`
			Documentation     string   `json:"documentation"`
			StoreCommand      string   `json:"storeCommand"`
			AutomationCommand string   `json:"automationCommand"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Data.DashboardSteps) < 4 {
		t.Fatalf("missing dashboard instructions: %#v", result.Data)
	}
	if result.Data.Documentation == "" ||
		result.Data.StoreCommand != "pb-agent connection add URL" ||
		!strings.Contains(result.Data.AutomationCommand, "--token-stdin") {
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

func TestConnectionAddHelpShowsSimpleUsage(t *testing.T) {
	var stdout bytes.Buffer
	code := Run([]string{"connection", "add", "--help"}, bytes.NewBuffer(nil), &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit code = %d, output = %s", code, stdout.String())
	}
	got := stdout.String()
	for _, expected := range []string{"pb-agent connection add URL", "--name", "--environment", "--token-stdin"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("help output missing %q:\n%s", expected, got)
		}
	}
	if strings.Contains(got, "--url") {
		t.Fatalf("help output contains removed --url flag:\n%s", got)
	}
}

func TestMutationHelpOnlyShowsApplicableFlags(t *testing.T) {
	tests := []struct {
		args       []string
		present    []string
		notPresent []string
	}{
		{[]string{"plan", "record-create", "--help"}, []string{"--collection", "--data-file"}, []string{"--id"}},
		{[]string{"plan", "record-delete", "--help"}, []string{"--collection", "--id"}, []string{"--data-file"}},
		{[]string{"plan", "collection-create", "--help"}, []string{"--data-file"}, []string{"--name"}},
		{[]string{"plan", "collection-delete", "--help"}, []string{"--name"}, []string{"--data-file"}},
	}
	for _, test := range tests {
		var stdout bytes.Buffer
		code := Run(test.args, bytes.NewBuffer(nil), &stdout, &bytes.Buffer{})
		if code != 0 {
			t.Fatalf("%v exit code = %d, output = %s", test.args, code, stdout.String())
		}
		for _, expected := range test.present {
			if !strings.Contains(stdout.String(), expected) {
				t.Errorf("%v help missing %q:\n%s", test.args, expected, stdout.String())
			}
		}
		for _, unexpected := range test.notPresent {
			if strings.Contains(stdout.String(), unexpected) {
				t.Errorf("%v help contains inapplicable %q:\n%s", test.args, unexpected, stdout.String())
			}
		}
	}
}

func TestConnectionAddBootstrapsConfigAndVerifiesToken(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/health":
			_, _ = w.Write([]byte(`{"message":"API is healthy."}`))
		case "/api/collections":
			_, _ = w.Write([]byte(`{"page":1,"perPage":1,"totalItems":0,"totalPages":0,"items":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, config.FileName)
	var savedReference, savedToken string
	var stdout bytes.Buffer
	a := &app{
		stdin:  strings.NewReader("secret-token\n"),
		stdout: &stdout,
		stderr: &bytes.Buffer{},
		saveCredential: func(reference, token string) error {
			savedReference, savedToken = reference, token
			return nil
		},
	}
	code := a.execute([]string{"--config", cfgPath, "connection", "add", server.URL, "--environment", "dev", "--token-stdin"})
	if code != 0 {
		t.Fatalf("exit code = %d, output = %s", code, stdout.String())
	}
	if authorization != "secret-token" {
		t.Fatalf("authorization = %q", authorization)
	}
	if savedReference != "default" || savedToken != "secret-token" {
		t.Fatalf("saved credential = %q, %q", savedReference, savedToken)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	connection := cfg.Connections["default"]
	if connection.URL != server.URL || connection.Environment != "development" {
		t.Fatalf("unexpected connection: %#v", connection)
	}
	var result struct {
		Data struct {
			Verified bool `json:"verified"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Data.Verified {
		t.Fatalf("connection was not reported as verified: %s", stdout.String())
	}
}

func TestConnectionAddDoesNotPersistUnverifiedToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/health" {
			_, _ = w.Write([]byte(`{"message":"API is healthy."}`))
			return
		}
		http.Error(w, `{"message":"authentication required"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	cfgPath := filepath.Join(t.TempDir(), config.FileName)
	saved := false
	var stdout bytes.Buffer
	a := &app{
		stdin:  strings.NewReader("bad-token\n"),
		stdout: &stdout,
		stderr: &bytes.Buffer{},
		saveCredential: func(_, _ string) error {
			saved = true
			return nil
		},
	}
	code := a.execute([]string{"--config", cfgPath, "connection", "add", server.URL, "--token-stdin"})
	if code == 0 {
		t.Fatalf("expected verification failure: %s", stdout.String())
	}
	if saved {
		t.Fatal("unverified credential was saved")
	}
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Fatalf("config should not exist after verification failure: %v", err)
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

func TestPlanApplyRecordUpsert(t *testing.T) {
	t.Setenv("PB_AGENT_STATE_DIR", t.TempDir())
	var upserted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/settings":
			_, _ = w.Write([]byte(`{"batch":{"enabled":true}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/collections/posts/records/fixedrecord0001":
			if upserted == nil {
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(upserted)
		case r.Method == http.MethodPost && r.URL.Path == "/api/batch":
			var batch struct {
				Requests []struct {
					Method string         `json:"method"`
					URL    string         `json:"url"`
					Body   map[string]any `json:"body"`
				} `json:"requests"`
			}
			_ = json.NewDecoder(r.Body).Decode(&batch)
			if len(batch.Requests) != 1 || batch.Requests[0].Method != http.MethodPut || batch.Requests[0].URL != "/api/collections/posts/records" {
				http.Error(w, `{"message":"invalid batch"}`, http.StatusBadRequest)
				return
			}
			upserted = batch.Requests[0].Body
			upserted["collectionName"] = "posts"
			_ = json.NewEncoder(w).Encode([]any{upserted})
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
	dataPath := filepath.Join(dir, "upsert.json")
	if err := os.WriteFile(dataPath, []byte(`{"name":"Created"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var planned bytes.Buffer
	code := Run([]string{"--config", cfgPath, "plan", "record-upsert", "--collection", "posts", "--id", "fixedrecord0001", "--data-file", dataPath}, bytes.NewBuffer(nil), &planned, &bytes.Buffer{})
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
	if upserted["id"] != "fixedrecord0001" || upserted["name"] != "Created" {
		t.Fatalf("record was not upserted: %#v", upserted)
	}
}

func TestRecordUpsertRequiresID(t *testing.T) {
	var stdout bytes.Buffer
	code := Run([]string{"plan", "record-upsert", "--collection", "posts", "--data-file", "ignored.json"}, bytes.NewBuffer(nil), &stdout, &bytes.Buffer{})
	if code != 2 || !strings.Contains(stdout.String(), "--id is required") {
		t.Fatalf("exit code = %d, output = %s", code, stdout.String())
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
