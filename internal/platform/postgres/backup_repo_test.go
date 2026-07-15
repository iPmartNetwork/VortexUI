package postgres

import (
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

func TestPreserveAdminSecretsFromRow(t *testing.T) {
	id := uuid.New()
	a := &domain.Admin{ID: id, Username: "admin", PasswordHash: "", TOTPSecret: "", WebhookSecret: ""}
	existing := db.Admin{
		ID: id, Username: "admin",
		PasswordHash: "$2a$12$existinghash", TotpSecret: "SECRET", WebhookSecret: "whsec",
	}
	preserveAdminSecretsFromRow(a, existing)
	if a.PasswordHash != existing.PasswordHash {
		t.Fatalf("password hash not preserved")
	}
	if a.TOTPSecret != existing.TotpSecret {
		t.Fatalf("totp secret not preserved")
	}
	if a.WebhookSecret != existing.WebhookSecret {
		t.Fatalf("webhook secret not preserved")
	}
}

func TestPreserveAdminSecretsFromRowDoesNotOverwrite(t *testing.T) {
	id := uuid.New()
	a := &domain.Admin{ID: id, PasswordHash: "newhash", TOTPSecret: "newtotp"}
	existing := db.Admin{PasswordHash: "oldhash", TotpSecret: "oldtotp"}
	preserveAdminSecretsFromRow(a, existing)
	if a.PasswordHash != "newhash" || a.TOTPSecret != "newtotp" {
		t.Fatalf("non-empty secrets should not be overwritten")
	}
}
