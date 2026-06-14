package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

func TestPasswordHashing(t *testing.T) {
	hash, err := HashPassword("s3cret-pass")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !CheckPassword(hash, "s3cret-pass") {
		t.Error("correct password rejected")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("wrong password accepted")
	}
}

func TestJWTRoundTrip(t *testing.T) {
	iss := NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	id := uuid.New()
	role := uuid.New()

	token, err := iss.Issue(id, false, &role)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := iss.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.AdminID != id || claims.Sudo || claims.RoleID == nil || *claims.RoleID != role {
		t.Errorf("claims round-trip mismatch: %+v", claims)
	}
}

func TestJWTRejectsTamperAndForeignSecret(t *testing.T) {
	iss := NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	other := NewIssuer([]byte("ffffffffffffffffffffffffffffffff"), time.Hour)
	token, _ := iss.Issue(uuid.New(), true, nil)

	if _, err := other.Verify(token); err == nil {
		t.Error("token verified under a different secret")
	}
	if _, err := iss.Verify(token + "x"); err == nil {
		t.Error("tampered token verified")
	}
}

func TestJWTRejectsExpired(t *testing.T) {
	iss := NewIssuer([]byte("0123456789abcdef0123456789abcdef"), -time.Minute) // already expired
	token, _ := iss.Issue(uuid.New(), true, nil)
	if _, err := iss.Verify(token); err == nil {
		t.Error("expired token accepted")
	}
}

func TestTOTP(t *testing.T) {
	secret, url, err := GenerateTOTP("VortexUI", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if secret == "" || url == "" {
		t.Fatal("empty secret/url")
	}
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("code: %v", err)
	}
	if !VerifyTOTP(secret, code) {
		t.Error("valid TOTP code rejected")
	}
	if VerifyTOTP(secret, "000000") {
		t.Error("bogus TOTP code accepted")
	}
}
