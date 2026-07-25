package audit

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/anirudh-777/pb-agent/internal/policy"
	"github.com/anirudh-777/pb-agent/internal/state"
)

type Event struct {
	SchemaVersion  string      `json:"schemaVersion"`
	ID             string      `json:"id"`
	Timestamp      time.Time   `json:"timestamp"`
	Connection     string      `json:"connection"`
	ConnectionHash string      `json:"connectionFingerprint"`
	Environment    string      `json:"environment"`
	Operation      string      `json:"operation"`
	Risk           policy.Risk `json:"risk"`
	PlanID         string      `json:"planId"`
	RequestHash    string      `json:"requestHash"`
	AffectedID     string      `json:"affectedId,omitempty"`
	ChangedFields  []string    `json:"changedFields"`
	Verified       bool        `json:"verified"`
	Outcome        string      `json:"outcome"`
	ErrorCode      string      `json:"errorCode,omitempty"`
}

func Append(event Event) (string, error) {
	rawID := make([]byte, 12)
	if _, err := rand.Read(rawID); err != nil {
		return "", err
	}
	event.SchemaVersion = "1"
	event.ID = "aud_" + hex.EncodeToString(rawID)
	event.Timestamp = time.Now().UTC()
	if event.ChangedFields == nil {
		event.ChangedFields = []string{}
	}
	dir, err := state.Dir()
	if err != nil {
		return "", err
	}
	file, err := os.OpenFile(filepath.Join(dir, "audit.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(event); err != nil {
		return "", err
	}
	return event.ID, file.Sync()
}

func HasRecentSuccess(connection, operation string, since time.Time) (bool, error) {
	dir, err := state.Dir()
	if err != nil {
		return false, err
	}
	file, err := os.Open(filepath.Join(dir, "audit.jsonl"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	found := false
	for scanner.Scan() {
		var event Event
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			continue
		}
		if event.Connection == connection && event.Operation == operation && event.Outcome == "succeeded" && !event.Timestamp.Before(since) {
			found = true
		}
	}
	return found, scanner.Err()
}
