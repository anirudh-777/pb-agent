package security

import (
	"encoding/json"
	"strings"
)

var secretFragments = []string{"password", "token", "authorization", "secret", "credential", "smtp", "encryption"}

func IsSecretKey(key string) bool {
	lower := strings.ToLower(key)
	for _, fragment := range secretFragments {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}

func Redact(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			if IsSecretKey(key) {
				out[key] = "[REDACTED]"
			} else {
				out[key] = Redact(child)
			}
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, child := range typed {
			out[i] = Redact(child)
		}
		return out
	default:
		return value
	}
}

func RedactJSON(raw []byte) any {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return "[UNPARSEABLE]"
	}
	return Redact(value)
}
