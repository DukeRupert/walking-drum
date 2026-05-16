package sqlc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

func TestSeasonLifecycle(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()

	// Season 1 was seeded by the 0004 migration as 'upcoming'.
	s1, err := q.GetSeasonByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetSeasonByID(1): %v", err)
	}
	if s1.Status != "upcoming" {
		t.Errorf("seeded season status: got %q, want upcoming", s1.Status)
	}

	if _, err := q.GetActiveSeason(ctx); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("GetActiveSeason before activation: got %v, want pgx.ErrNoRows", err)
	}

	activated, err := q.UpdateSeasonStatus(ctx, sqlc.UpdateSeasonStatusParams{
		ID:     1,
		Status: "active",
	})
	if err != nil {
		t.Fatalf("UpdateSeasonStatus -> active: %v", err)
	}
	if activated.Status != "active" {
		t.Errorf("status after activation: got %q, want active", activated.Status)
	}

	active, err := q.GetActiveSeason(ctx)
	if err != nil {
		t.Fatalf("GetActiveSeason after activation: %v", err)
	}
	if active.ID != 1 {
		t.Errorf("active season id: got %d, want 1", active.ID)
	}

	ended, err := q.UpdateSeasonStatus(ctx, sqlc.UpdateSeasonStatusParams{
		ID:     1,
		Status: "ended",
	})
	if err != nil {
		t.Fatalf("UpdateSeasonStatus -> ended: %v", err)
	}
	if ended.Status != "ended" {
		t.Errorf("status after ending: got %q, want ended", ended.Status)
	}
}

func TestOnlyOneActiveSeason(t *testing.T) {
	q, tx := testdb.WithTx(t)
	ctx := context.Background()

	// Insert a second 'upcoming' season alongside the seeded one.
	if _, err := tx.Exec(ctx, `
		INSERT INTO seasons (id, name, status, world_seed, starts_at, ends_at)
		VALUES (2, 'Season 2', 'upcoming', 0, NOW(), NOW() + interval '90 days')
	`); err != nil {
		t.Fatalf("insert season 2: %v", err)
	}

	if _, err := q.UpdateSeasonStatus(ctx, sqlc.UpdateSeasonStatusParams{
		ID: 1, Status: "active",
	}); err != nil {
		t.Fatalf("activate season 1: %v", err)
	}

	if _, err := q.UpdateSeasonStatus(ctx, sqlc.UpdateSeasonStatusParams{
		ID: 2, Status: "active",
	}); err == nil {
		t.Fatal("expected unique-violation activating a second season, got nil")
	}
}
