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

func appendModeration(t *testing.T, ctx context.Context, q *sqlc.Queries, accID pgtype.UUID, actionType, reason string, expires *time.Time) sqlc.ModerationAction {
	t.Helper()
	id, _ := uuid.NewV7()
	exp := pgtype.Timestamptz{}
	if expires != nil {
		exp = pgtype.Timestamptz{Time: *expires, Valid: true}
	}
	row, err := q.AppendModerationAction(ctx, sqlc.AppendModerationActionParams{
		ID:         pgtype.UUID{Bytes: id, Valid: true},
		AccountID:  accID,
		ActionType: actionType,
		Reason:     reason,
		Details:    []byte("{}"),
		ExpiresAt:  exp,
	})
	if err != nil {
		t.Fatalf("AppendModerationAction(%s): %v", actionType, err)
	}
	return row
}

func TestModerationAppendAndList(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	acc := makeAccount(t, ctx, q, "mod-test@example.com", "ModTest")

	appendModeration(t, ctx, q, acc.ID, "warn", "language", nil)
	appendModeration(t, ctx, q, acc.ID, "mute", "spamming chat", nil)

	rows, err := q.ListModerationActionsForAccount(ctx, acc.ID)
	if err != nil {
		t.Fatalf("ListModerationActionsForAccount: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows: got %d, want 2", len(rows))
	}
	// DESC by applied_at — newest first.
	if rows[0].ActionType != "mute" || rows[1].ActionType != "warn" {
		t.Errorf("ordering: got [%s, %s], want [mute, warn]", rows[0].ActionType, rows[1].ActionType)
	}
}

func TestModerationActionTypeCheck(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	acc := makeAccount(t, ctx, q, "mod-check@example.com", "ModCheck")

	id, _ := uuid.NewV7()
	if _, err := q.AppendModerationAction(ctx, sqlc.AppendModerationActionParams{
		ID:         pgtype.UUID{Bytes: id, Valid: true},
		AccountID:  acc.ID,
		ActionType: "yeet",
		Reason:     "n/a",
		Details:    []byte("{}"),
	}); err == nil {
		t.Fatal("expected CHECK violation for bogus action_type")
	}
}

func TestFindActiveBansAndSuspensions(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	acc := makeAccount(t, ctx, q, "ban-test@example.com", "BanTest")

	// 1. Permanent ban -> shows as active.
	appendModeration(t, ctx, q, acc.ID, "ban", "cheating", nil)
	active, err := q.FindActiveBansAndSuspensions(ctx, acc.ID)
	if err != nil {
		t.Fatalf("FindActive: %v", err)
	}
	if len(active) != 1 || active[0].ActionType != "ban" {
		t.Fatalf("after ban: got %d rows, want 1 ban", len(active))
	}

	// 2. Subsequent unban -> active list empties.
	appendModeration(t, ctx, q, acc.ID, "unban", "appeal granted", nil)
	active, err = q.FindActiveBansAndSuspensions(ctx, acc.ID)
	if err != nil {
		t.Fatalf("FindActive after unban: %v", err)
	}
	if len(active) != 0 {
		t.Errorf("after unban: got %d rows, want 0", len(active))
	}

	// 3. Expired suspension -> not in active list.
	past := time.Now().Add(-time.Hour)
	appendModeration(t, ctx, q, acc.ID, "suspend", "tilt", &past)
	active, err = q.FindActiveBansAndSuspensions(ctx, acc.ID)
	if err != nil {
		t.Fatalf("FindActive after expired suspend: %v", err)
	}
	if len(active) != 0 {
		t.Errorf("after expired suspend: got %d rows, want 0", len(active))
	}

	// 4. Live suspension -> active again.
	future := time.Now().Add(time.Hour)
	appendModeration(t, ctx, q, acc.ID, "suspend", "still tilt", &future)
	active, err = q.FindActiveBansAndSuspensions(ctx, acc.ID)
	if err != nil {
		t.Fatalf("FindActive after live suspend: %v", err)
	}
	if len(active) != 1 || active[0].ActionType != "suspend" {
		t.Errorf("after live suspend: got %d rows (want 1 suspend)", len(active))
	}

	// 5. Warns and mutes are ignored by this query.
	appendModeration(t, ctx, q, acc.ID, "warn", "language", nil)
	appendModeration(t, ctx, q, acc.ID, "mute", "spam", nil)
	active, err = q.FindActiveBansAndSuspensions(ctx, acc.ID)
	if err != nil {
		t.Fatalf("FindActive after warn/mute: %v", err)
	}
	if len(active) != 1 {
		t.Errorf("warn/mute should not affect active bans+suspends list; got %d, want 1", len(active))
	}
}
