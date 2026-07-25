package policy

import (
	"testing"
	"time"
)

func TestProductionWriteRequiresMatchingGrant(t *testing.T) {
	now := time.Now()
	if err := ValidateGrant(nil, "prod", "production", "records.write", now); err == nil {
		t.Fatal("expected missing grant error")
	}
	grant := &Grant{Connection: "prod", Environment: "production", Scope: "records.write", ExpiresAt: now.Add(time.Minute)}
	if err := ValidateGrant(grant, "prod", "production", "records.write", now); err != nil {
		t.Fatal(err)
	}
}
