package security

import "testing"

func TestRedactNestedSecrets(t *testing.T) {
	input := map[string]any{
		"name": "safe",
		"nested": map[string]any{
			"password": "bad",
			"apiToken": "bad",
		},
	}
	got := Redact(input).(map[string]any)
	nested := got["nested"].(map[string]any)
	if nested["password"] != "[REDACTED]" || nested["apiToken"] != "[REDACTED]" {
		t.Fatalf("secrets were not redacted: %#v", got)
	}
	if got["name"] != "safe" {
		t.Fatalf("safe value changed: %#v", got)
	}
}
