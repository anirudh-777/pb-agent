package policy

import (
	"fmt"
	"time"
)

type Risk string

const (
	Read        Risk = "read"
	Write       Risk = "write"
	Destructive Risk = "destructive"
	Privileged  Risk = "privileged"
)

type Grant struct {
	Connection  string    `json:"connection"`
	Environment string    `json:"environment"`
	Scope       string    `json:"scope"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func RequiresGrant(environment string, risk Risk) bool {
	if risk == Read {
		return false
	}
	return environment == "staging" || environment == "production"
}

func ValidateGrant(grant *Grant, connection, environment, scope string, now time.Time) error {
	if !RequiresGrant(environment, Write) {
		return nil
	}
	if grant == nil {
		return fmt.Errorf("a temporary access grant is required")
	}
	if grant.Connection != connection || grant.Environment != environment {
		return fmt.Errorf("access grant targets a different connection")
	}
	if grant.Scope != scope && grant.Scope != "*" {
		return fmt.Errorf("access grant does not include scope %q", scope)
	}
	if !now.Before(grant.ExpiresAt) {
		return fmt.Errorf("access grant has expired")
	}
	return nil
}
