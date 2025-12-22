package auth

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// Principal holds identity extracted from a validated token.
type Principal struct {
	UserID string
	Roles  []string
	OrgID  string
	Claims jwt.MapClaims
}

var (
	ErrNoToken       = errors.New("no token provided")
	ErrInvalidToken  = errors.New("invalid token")
	ErrInvalidIssuer = errors.New("invalid issuer")
	ErrMissingSub    = errors.New("missing sub claim")
)

type verifier struct {
	cfg  Config
	jwks *JWKS
}

// NewVerifier constructs a verifier with config and JWKS.
func NewVerifier(cfg Config, jwks *JWKS) *verifier {
	return &verifier{cfg: cfg, jwks: jwks}
}

// ParseAndVerifyToken verifies a bearer token, validates issuer/exp and returns Principal.
func (v *verifier) ParseAndVerifyToken(tokenString string) (*Principal, error) {
	if tokenString == "" {
		return nil, ErrNoToken
	}
	tokenString = strings.TrimSpace(tokenString)
	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// enforce RS256
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, ErrInvalidToken
		}
		return v.jwks.Get(kid)
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	// issuer
	if iss, _ := claims["iss"].(string); iss != v.cfg.Issuer {
		return nil, ErrInvalidIssuer
	}
	// exp
	if !claims.VerifyExpiresAt(jwt.TimeFunc().Unix(), true) {
		return nil, ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, ErrMissingSub
	}

	// extract roles from realm_access.roles
	var roles []string
	if ra, ok := claims["realm_access"].(map[string]interface{}); ok {
		if rr, ok := ra["roles"].([]interface{}); ok {
			for _, r := range rr {
				if s, ok := r.(string); ok {
					roles = append(roles, s)
				}
			}
		}
	}

	// organisationId may be string or number
	var orgID string
	if v, ok := claims["organisationId"].(string); ok {
		orgID = v
	}

	return &Principal{
		UserID: sub,
		Roles:  roles,
		OrgID:  orgID,
		Claims: claims,
	}, nil
}
