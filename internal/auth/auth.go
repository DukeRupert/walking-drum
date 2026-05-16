// Package auth holds the data-layer auth primitives: bcrypt password hashing,
// session-token generation, and the helpers that connect them to the
// sessions/accounts tables. There is no HTTP code here — HTTP handlers are
// a separate, later concern (see TODO.md "Explicitly Not in This List").
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
)

// HashPassword bcrypts a plaintext password; the returned string is what
// goes into accounts.password_hash.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt: %w", err)
	}
	return string(b), nil
}

// VerifyPassword returns nil if plain matches hash, or an error otherwise.
// The bcrypt-returned error preserves "mismatched hash and password" so
// callers can decide whether to leak that fact (typically: don't).
func VerifyPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

const tokenRandBytes = 32

// GenerateSessionToken returns (rawToken, tokenHash). The raw token is
// what gets handed to the client exactly once; the hash is what we
// store in sessions.token_hash. Per DESIGN.md §5.5 the raw token
// must not be persisted server-side.
func GenerateSessionToken() (raw, hash string, err error) {
	buf := make([]byte, tokenRandBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("rand: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, HashToken(raw), nil
}

// HashToken is the one-way function used both to store and to look up
// session tokens. SHA-256 (not bcrypt) because session tokens are
// already high-entropy 32-byte random values; bcrypting them burns CPU
// without measurably improving security.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// DefaultSessionTTL is the lifetime applied to a fresh session when the
// caller doesn't override it.
const DefaultSessionTTL = 30 * 24 * time.Hour

// CreateSessionForAccount provisions a fresh session row for account and
// returns the raw token. The raw token is the only thing the client ever
// sees; the database only ever sees its hash.
func CreateSessionForAccount(ctx context.Context, q *sqlc.Queries, accountID pgtype.UUID, ttl time.Duration) (rawToken string, sess sqlc.Session, err error) {
	if ttl <= 0 {
		ttl = DefaultSessionTTL
	}
	raw, hash, err := GenerateSessionToken()
	if err != nil {
		return "", sqlc.Session{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", sqlc.Session{}, fmt.Errorf("session id: %w", err)
	}
	sess, err = q.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		AccountID: accountID,
		TokenHash: hash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(ttl), Valid: true},
	})
	if err != nil {
		return "", sqlc.Session{}, fmt.Errorf("insert session: %w", err)
	}
	return raw, sess, nil
}

var (
	ErrSessionNotFound = errors.New("auth: session not found")
	ErrSessionRevoked  = errors.New("auth: session revoked")
	ErrSessionExpired  = errors.New("auth: session expired")
)

// ValidateSessionToken looks up a session by the raw token and returns
// the row only if it is neither revoked nor expired. It does not touch
// last_seen_at — that's a separate concern for callers who want it.
func ValidateSessionToken(ctx context.Context, q *sqlc.Queries, rawToken string) (sqlc.Session, error) {
	sess, err := q.GetSessionByTokenHash(ctx, HashToken(rawToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Session{}, ErrSessionNotFound
		}
		return sqlc.Session{}, err
	}
	if sess.RevokedAt.Valid {
		return sqlc.Session{}, ErrSessionRevoked
	}
	if !sess.ExpiresAt.Valid || !sess.ExpiresAt.Time.After(time.Now()) {
		return sqlc.Session{}, ErrSessionExpired
	}
	return sess, nil
}
