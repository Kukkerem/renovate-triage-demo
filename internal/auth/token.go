// Package auth issues and verifies short-lived service tokens used for
// provider-to-provider calls (e.g. exchanging a technical user's credentials
// for a scoped token before calling the BTP XSUAA-protected APIs).
//
// Tokens are HS256-signed and short-lived; ParseServiceToken restricts the
// accepted algorithms explicitly rather than trusting the header.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// ServiceClaims is the claim set issued for service-to-service auth between
// this provider and a downstream BTP API.
type ServiceClaims struct {
	jwt.StandardClaims
	Scope []string `json:"scope"`
}

// strictParser only accepts HMAC-SHA256 tokens and never silently skips
// exp/nbf validation. In jwt/v4 both knobs are plain exported struct fields;
// in jwt/v5 the same struct's fields are unexported and set via
// jwt.NewParser(jwt.WithValidMethods(...)) instead.
var strictParser = &jwt.Parser{
	ValidMethods:         []string{"HS256"},
	SkipClaimsValidation: false,
}

// IssueToken signs a short-lived service token carrying the given scopes.
func IssueToken(secret []byte, subject string, scopes []string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := ServiceClaims{
		StandardClaims: jwt.StandardClaims{
			Subject:   subject,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(ttl).Unix(),
		},
		Scope: scopes,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("auth: sign token: %w", err)
	}
	return signed, nil
}

// VerifyToken parses and validates a service token, returning its claims.
func VerifyToken(secret []byte, raw string) (*ServiceClaims, error) {
	claims := &ServiceClaims{}
	token, err := strictParser.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("auth: parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("auth: token failed validation")
	}
	// StandardClaims.Valid() checks exp/iat/nbf. In jwt/v5 the base Claims
	// interface no longer has a Valid() method at all (it is replaced by
	// getter methods), so this call site is doubly broken there.
	if err := claims.Valid(); err != nil {
		return nil, fmt.Errorf("auth: claims invalid: %w", err)
	}
	return claims, nil
}
