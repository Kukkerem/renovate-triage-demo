package auth

import (
	"strings"
	"testing"
	"time"
)

func TestIssueAndVerifyToken_RoundTrip(t *testing.T) {
	secret := []byte("test-signing-key")
	scopes := []string{"btp.read", "btp.write"}

	signed, err := IssueToken(secret, "provider@demo", scopes, time.Hour)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if signed == "" {
		t.Fatal("IssueToken returned an empty token string")
	}

	claims, err := VerifyToken(secret, signed)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if claims.Subject != "provider@demo" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "provider@demo")
	}
	if len(claims.Scope) != len(scopes) {
		t.Fatalf("Scope = %v, want %v", claims.Scope, scopes)
	}
	for i, s := range scopes {
		if claims.Scope[i] != s {
			t.Errorf("Scope[%d] = %q, want %q", i, claims.Scope[i], s)
		}
	}
}

func TestVerifyToken_RejectsWrongSecret(t *testing.T) {
	signed, err := IssueToken([]byte("right-secret"), "provider@demo", nil, time.Hour)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if _, err := VerifyToken([]byte("wrong-secret"), signed); err == nil {
		t.Fatal("VerifyToken succeeded with the wrong secret, want error")
	}
}

func TestVerifyToken_RejectsExpiredToken(t *testing.T) {
	secret := []byte("test-signing-key")
	signed, err := IssueToken(secret, "provider@demo", nil, -time.Minute)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	_, err = VerifyToken(secret, signed)
	if err == nil {
		t.Fatal("VerifyToken succeeded on an expired token, want error")
	}
	if !strings.Contains(err.Error(), "expired") && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("unexpected error for expired token: %v", err)
	}
}
