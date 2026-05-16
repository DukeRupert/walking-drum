package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dukerupert/walking-drum/internal/auth"
	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := auth.HashPassword("hunter2")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "hunter2" {
		t.Fatal("hash must not equal plaintext")
	}
	if err := auth.VerifyPassword(hash, "hunter2"); err != nil {
		t.Errorf("VerifyPassword(correct): %v", err)
	}
	if err := auth.VerifyPassword(hash, "wrong-password"); err == nil {
		t.Error("VerifyPassword(wrong) should fail")
	}
}

func TestSessionTokenIsRandom(t *testing.T) {
	a, hashA, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken: %v", err)
	}
	b, hashB, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken: %v", err)
	}
	if a == b {
		t.Fatal("two calls produced the same raw token — randomness broken")
	}
	if hashA == hashB {
		t.Fatal("two calls produced the same token hash")
	}
	if auth.HashToken(a) != hashA {
		t.Error("HashToken is not deterministic for the same input")
	}
}

// The full Phase 1 "Done when" demonstration: create an account,
// create a session, validate the session token, revoke it.
func TestAuthRoundTrip(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()

	pwHash, err := auth.HashPassword("hunter2")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	accID, _ := uuid.NewV7()
	acc, err := q.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:           pgtype.UUID{Bytes: accID, Valid: true},
		Email:        "roundtrip@example.com",
		DisplayName:  "RoundTrip",
		PasswordHash: pwHash,
	})
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}

	rawToken, sess, err := auth.CreateSessionForAccount(ctx, q, acc.ID, 0)
	if err != nil {
		t.Fatalf("CreateSessionForAccount: %v", err)
	}
	if rawToken == "" {
		t.Fatal("raw token should be non-empty")
	}
	if sess.TokenHash == rawToken {
		t.Fatal("stored TokenHash must not equal raw token")
	}

	validated, err := auth.ValidateSessionToken(ctx, q, rawToken)
	if err != nil {
		t.Fatalf("ValidateSessionToken: %v", err)
	}
	if validated.ID != sess.ID {
		t.Errorf("validated session id mismatch")
	}

	// Wrong token must not validate.
	if _, err := auth.ValidateSessionToken(ctx, q, "not-the-real-token"); !errors.Is(err, auth.ErrSessionNotFound) {
		t.Errorf("bogus token: got %v, want ErrSessionNotFound", err)
	}

	// Revoke and re-validate.
	reason := "user_logout"
	if _, err := q.RevokeSession(ctx, sqlc.RevokeSessionParams{
		ID:           sess.ID,
		RevokeReason: &reason,
	}); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}
	if _, err := auth.ValidateSessionToken(ctx, q, rawToken); !errors.Is(err, auth.ErrSessionRevoked) {
		t.Errorf("post-revoke validation: got %v, want ErrSessionRevoked", err)
	}
}
