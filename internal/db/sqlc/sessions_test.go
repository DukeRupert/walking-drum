package sqlc_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

// makeAccount inserts a fresh account into the current tx and returns its row,
// so tests that depend on a real account_id can share one helper.
func makeAccount(t *testing.T, ctx context.Context, q *sqlc.Queries, email, name string) sqlc.Account {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("uuid: %v", err)
	}
	acc, err := q.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		Email:        email,
		DisplayName:  name,
		PasswordHash: "x",
	})
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	return acc
}

func TestSessionsLifecycle(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	acc := makeAccount(t, ctx, q, "session-test@example.com", "SessionTest")

	sessionID, _ := uuid.NewV7()
	expires := time.Now().Add(24 * time.Hour)
	sess, err := q.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:        pgtype.UUID{Bytes: sessionID, Valid: true},
		AccountID: acc.ID,
		TokenHash: "fake-token-hash-1",
		ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.RevokedAt.Valid {
		t.Error("freshly-created session should not be revoked")
	}

	byHash, err := q.GetSessionByTokenHash(ctx, "fake-token-hash-1")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash: %v", err)
	}
	if byHash.ID != sess.ID {
		t.Errorf("token-hash lookup returned wrong session id")
	}

	active, err := q.ListActiveSessionsForAccount(ctx, acc.ID)
	if err != nil {
		t.Fatalf("ListActiveSessionsForAccount: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("active sessions: got %d, want 1", len(active))
	}

	reason := "user_logout"
	revoked, err := q.RevokeSession(ctx, sqlc.RevokeSessionParams{
		ID:           sess.ID,
		RevokeReason: &reason,
	})
	if err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}
	if !revoked.RevokedAt.Valid {
		t.Error("RevokedAt should be set after RevokeSession")
	}

	active, err = q.ListActiveSessionsForAccount(ctx, acc.ID)
	if err != nil {
		t.Fatalf("ListActiveSessionsForAccount after revoke: %v", err)
	}
	if len(active) != 0 {
		t.Errorf("active sessions after revoke: got %d, want 0", len(active))
	}
}

func TestSessionRevokeReasonCheck(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	acc := makeAccount(t, ctx, q, "check-test@example.com", "CheckTest")

	sessionID, _ := uuid.NewV7()
	sess, err := q.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:        pgtype.UUID{Bytes: sessionID, Valid: true},
		AccountID: acc.ID,
		TokenHash: "fake-token-hash-2",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	bogus := "not-a-real-reason"
	if _, err := q.RevokeSession(ctx, sqlc.RevokeSessionParams{
		ID:           sess.ID,
		RevokeReason: &bogus,
	}); err == nil {
		t.Fatal("expected CHECK constraint violation for bogus revoke_reason")
	}
}

func TestSessionExpiredIsNotActive(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	acc := makeAccount(t, ctx, q, "expired-test@example.com", "ExpiredTest")

	sessionID, _ := uuid.NewV7()
	if _, err := q.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:        pgtype.UUID{Bytes: sessionID, Valid: true},
		AccountID: acc.ID,
		TokenHash: "fake-token-hash-3",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
	}); err != nil {
		t.Fatalf("CreateSession (expired): %v", err)
	}

	active, err := q.ListActiveSessionsForAccount(ctx, acc.ID)
	if err != nil {
		t.Fatalf("ListActiveSessionsForAccount: %v", err)
	}
	if len(active) != 0 {
		t.Errorf("active sessions: got %d, want 0 (session is expired)", len(active))
	}
}
