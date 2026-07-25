package plan

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anirudh-777/pb-agent/internal/policy"
	"github.com/anirudh-777/pb-agent/internal/state"
)

const lifetime = 15 * time.Minute

type Plan struct {
	SchemaVersion    string      `json:"schemaVersion"`
	ID               string      `json:"id"`
	Operation        string      `json:"operation"`
	Connection       string      `json:"connection"`
	ConnectionHash   string      `json:"connectionFingerprint"`
	Environment      string      `json:"environment"`
	Risk             policy.Risk `json:"risk"`
	Scope            string      `json:"scope"`
	Method           string      `json:"method"`
	Path             string      `json:"path"`
	EncryptedPayload string      `json:"encryptedPayload,omitempty"`
	RequestHash      string      `json:"requestHash"`
	PreconditionPath string      `json:"preconditionPath,omitempty"`
	PreconditionHash string      `json:"preconditionHash,omitempty"`
	Preview          any         `json:"preview"`
	CreatedAt        time.Time   `json:"createdAt"`
	ExpiresAt        time.Time   `json:"expiresAt"`
	AppliedAt        *time.Time  `json:"appliedAt,omitempty"`
	Integrity        string      `json:"integrity"`
}

type Public struct {
	SchemaVersion    string      `json:"schemaVersion"`
	ID               string      `json:"id"`
	Operation        string      `json:"operation"`
	Connection       string      `json:"connection"`
	Environment      string      `json:"environment"`
	Risk             policy.Risk `json:"risk"`
	Scope            string      `json:"scope"`
	RequestHash      string      `json:"requestHash"`
	PreconditionHash string      `json:"preconditionHash,omitempty"`
	Preview          any         `json:"preview"`
	CreatedAt        time.Time   `json:"createdAt"`
	ExpiresAt        time.Time   `json:"expiresAt"`
	RequiresGrant    bool        `json:"requiresGrant"`
}

func (p Plan) Public() Public {
	return Public{
		SchemaVersion:    p.SchemaVersion,
		ID:               p.ID,
		Operation:        p.Operation,
		Connection:       p.Connection,
		Environment:      p.Environment,
		Risk:             p.Risk,
		Scope:            p.Scope,
		RequestHash:      p.RequestHash,
		PreconditionHash: p.PreconditionHash,
		Preview:          p.Preview,
		CreatedAt:        p.CreatedAt,
		ExpiresAt:        p.ExpiresAt,
		RequiresGrant:    policy.RequiresGrant(p.Environment, p.Risk),
	}
}

func New(operation, connection, connectionHash, environment string, risk policy.Risk, scope, method, path string, payload []byte, preview any, preconditionPath, preconditionHash string, now time.Time) (Plan, error) {
	id, err := randomID("pln_", 12)
	if err != nil {
		return Plan{}, err
	}
	encrypted, err := encrypt(payload)
	if err != nil {
		return Plan{}, err
	}
	sum := sha256.Sum256(append([]byte(method+"\n"+path+"\n"), payload...))
	return Plan{
		SchemaVersion:    "1",
		ID:               id,
		Operation:        operation,
		Connection:       connection,
		ConnectionHash:   connectionHash,
		Environment:      environment,
		Risk:             risk,
		Scope:            scope,
		Method:           method,
		Path:             path,
		EncryptedPayload: encrypted,
		RequestHash:      hex.EncodeToString(sum[:]),
		PreconditionPath: preconditionPath,
		PreconditionHash: preconditionHash,
		Preview:          preview,
		CreatedAt:        now.UTC(),
		ExpiresAt:        now.UTC().Add(lifetime),
	}, nil
}

func Save(plan Plan) error {
	if err := state.SafeID(plan.ID); err != nil {
		return err
	}
	dir, err := state.PlansDir()
	if err != nil {
		return err
	}
	integrity, err := sign(plan)
	if err != nil {
		return err
	}
	plan.Integrity = integrity
	raw, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, plan.ID+".json")
	temp := path + ".tmp"
	if err := os.WriteFile(temp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func Load(id string) (Plan, error) {
	if err := state.SafeID(id); err != nil {
		return Plan{}, err
	}
	dir, err := state.PlansDir()
	if err != nil {
		return Plan{}, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return Plan{}, err
	}
	var result Plan
	if err := json.Unmarshal(raw, &result); err != nil {
		return Plan{}, err
	}
	expected, err := sign(result)
	if err != nil {
		return Plan{}, err
	}
	if !hmac.Equal([]byte(result.Integrity), []byte(expected)) {
		return Plan{}, errors.New("plan integrity check failed")
	}
	return result, nil
}

func (p Plan) Validate(connectionHash, environment string, now time.Time) error {
	if p.SchemaVersion != "1" {
		return fmt.Errorf("unsupported plan schema")
	}
	if p.ConnectionHash != connectionHash {
		return fmt.Errorf("plan targets a different PocketBase instance")
	}
	if p.Environment != environment {
		return fmt.Errorf("connection environment changed after planning")
	}
	if p.AppliedAt != nil {
		return fmt.Errorf("plan was already applied")
	}
	if !now.Before(p.ExpiresAt) {
		return fmt.Errorf("plan has expired")
	}
	return nil
}

func Acquire(id string) (func() error, error) {
	if err := state.SafeID(id); err != nil {
		return nil, err
	}
	dir, err := state.PlansDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, id+".lock")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if os.IsExist(err) {
		return nil, errors.New("plan is already being applied or requires manual lock recovery")
	}
	if err != nil {
		return nil, err
	}
	_, writeErr := fmt.Fprintf(file, "pid=%d\ncreated=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
	closeErr := file.Close()
	if writeErr != nil {
		_ = os.Remove(path)
		return nil, writeErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return nil, closeErr
	}
	return func() error { return os.Remove(path) }, nil
}

func (p Plan) Payload() ([]byte, error) {
	raw, err := decrypt(p.EncryptedPayload)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(append([]byte(p.Method+"\n"+p.Path+"\n"), raw...))
	if hex.EncodeToString(sum[:]) != p.RequestHash {
		return nil, errors.New("plan payload hash mismatch")
	}
	return raw, nil
}

func Hash(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func randomID(prefix string, size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(raw), nil
}

func encrypt(plaintext []byte) (string, error) {
	if len(plaintext) == 0 {
		return "", nil
	}
	key, err := localKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, nil
	}
	key, err := localKey()
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("invalid encrypted plan")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	return gcm.Open(nil, nonce, ciphertext[gcm.NonceSize():], nil)
}

func localKey() ([]byte, error) {
	dir, err := state.Dir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "plan.key")
	if raw, err := os.ReadFile(path); err == nil {
		if len(raw) != 32 {
			return nil, errors.New("invalid local plan key")
		}
		return raw, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	temp, err := os.CreateTemp(dir, "plan-key-*")
	if err != nil {
		return nil, err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return nil, err
	}
	if _, err := temp.Write(raw); err != nil {
		_ = temp.Close()
		return nil, err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return nil, err
	}
	if err := temp.Close(); err != nil {
		return nil, err
	}
	if err := os.Link(tempPath, path); err != nil && !os.IsExist(err) {
		return nil, err
	}
	return os.ReadFile(path)
}

func sign(value Plan) (string, error) {
	value.Integrity = ""
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	key, err := localKey()
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(raw)
	return hex.EncodeToString(mac.Sum(nil)), nil
}
