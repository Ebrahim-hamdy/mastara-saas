package security

import (
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/config"
	"github.com/google/uuid"
)

// AuthPayload contains the data embedded within an authentication token.
type AuthPayload struct {
	TokenID     uuid.UUID   `json:"jti"`
	UserID      uuid.UUID   `json:"uid"`
	ClinicID    uuid.UUID   `json:"cid"`
	RoleIDs     []uuid.UUID `json:"roles"`
	Permissions []string    `json:"perms"`
	IssuedAt    time.Time   `json:"iat"`
	ExpiresAt   time.Time   `json:"exp"`
}

// NewAuthPayload creates a new payload for a user token.
func NewAuthPayload(userID, clinicID uuid.UUID, roleIDs []uuid.UUID, permissions []string, duration time.Duration) (*AuthPayload, error) {
	tokenID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token ID: %w", err)
	}

	now := time.Now().UTC()
	payload := &AuthPayload{
		TokenID:     tokenID,
		UserID:      userID,
		ClinicID:    clinicID,
		RoleIDs:     roleIDs,
		Permissions: permissions,
		IssuedAt:    now,
		ExpiresAt:   now.Add(duration),
	}
	return payload, nil
}

// IsValid checks if the token payload has expired.
func (p *AuthPayload) IsValid() error {
	if time.Now().UTC().After(p.ExpiresAt) {
		return fmt.Errorf("token has expired")
	}
	return nil
}

// PasetoManager is a PASETO token manager using the aidantwoods/go-paseto library.
type PasetoManager struct {
	symmetricKey paseto.V4SymmetricKey
}

// NewPasetoManager creates a new PasetoManager.
func NewPasetoManager(cfg config.SecurityConfig) (*PasetoManager, error) {
	if len(cfg.PasetoKey) != 32 {
		return nil, fmt.Errorf("invalid paseto key size: must be exactly 32 characters")
	}

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(cfg.PasetoKey))
	if err != nil {
		return nil, fmt.Errorf("failed to construct paseto symmetric key: %w", err)
	}

	return &PasetoManager{
		symmetricKey: key,
	}, nil
}

// CreateToken creates a new PASETO v4.local token for a given payload.
func (m *PasetoManager) CreateToken(payload *AuthPayload) (string, error) {
	token := paseto.NewToken()
	token.SetJti(payload.TokenID.String())
	token.SetIssuedAt(payload.IssuedAt)
	token.SetExpiration(payload.ExpiresAt)

	// --- THIS IS THE CRITICAL CORRECTION ---
	// SetString and Set do not return errors.
	token.SetString("uid", payload.UserID.String())
	token.SetString("cid", payload.ClinicID.String())
	token.Set("roles", payload.RoleIDs)
	token.Set("perms", payload.Permissions)

	// V4Encrypt returns a single string value.
	encryptedToken := token.V4Encrypt(m.symmetricKey, nil)
	return encryptedToken, nil
	// --- END CORRECTION ---
}

// VerifyToken checks if the token is valid and returns its payload.
func (m *PasetoManager) VerifyToken(tokenString string) (*AuthPayload, error) {
	parser := paseto.NewParser()
	token, err := parser.ParseV4Local(m.symmetricKey, tokenString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse or validate token: %w", err)
	}

	payload := &AuthPayload{}

	jtiStr, err := token.GetJti()
	if err != nil {
		return nil, fmt.Errorf("failed to get token jti: %w", err)
	}
	if payload.TokenID, err = uuid.Parse(jtiStr); err != nil {
		return nil, fmt.Errorf("invalid jti in token: %w", err)
	}

	payload.IssuedAt, err = token.GetIssuedAt()
	if err != nil {
		return nil, fmt.Errorf("failed to get token iat: %w", err)
	}

	payload.ExpiresAt, err = token.GetExpiration()
	if err != nil {
		return nil, fmt.Errorf("failed to get token exp: %w", err)
	}

	userIDStr, err := token.GetString("uid")
	if err != nil {
		return nil, fmt.Errorf("failed to get user id from token: %w", err)
	}
	if payload.UserID, err = uuid.Parse(userIDStr); err != nil {
		return nil, fmt.Errorf("invalid user id in token: %w", err)
	}

	clinicIDStr, err := token.GetString("cid")
	if err != nil {
		return nil, fmt.Errorf("failed to get clinic id from token: %w", err)
	}
	if payload.ClinicID, err = uuid.Parse(clinicIDStr); err != nil {
		return nil, fmt.Errorf("invalid clinic id in token: %w", err)
	}

	if err := token.Get("roles", &payload.RoleIDs); err != nil {
		return nil, fmt.Errorf("failed to get roles from token: %w", err)
	}
	if err := token.Get("perms", &payload.Permissions); err != nil {
		return nil, fmt.Errorf("failed to get permissions from token: %w", err)
	}

	if err := payload.IsValid(); err != nil {
		return nil, err
	}

	return payload, nil
}
