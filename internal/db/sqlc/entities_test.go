package sqlc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

func newEntityID(t *testing.T) pgtype.UUID {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("uuid: %v", err)
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func TestEntitiesCRUD(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()

	id := newEntityID(t)
	ent, err := q.CreateEntity(ctx, sqlc.CreateEntityParams{
		ID:            id,
		SeasonID:      1, // seeded by 00004_seasons.sql
		EntityType:    "character",
		CreatedAtTick: 100,
	})
	if err != nil {
		t.Fatalf("CreateEntity: %v", err)
	}
	if ent.EntityType != "character" {
		t.Errorf("entity_type: got %q, want character", ent.EntityType)
	}
	if ent.DestroyedAtTick != nil {
		t.Errorf("destroyed_at_tick on fresh row: got %v, want nil", ent.DestroyedAtTick)
	}

	got, err := q.GetEntityByID(ctx, id)
	if err != nil {
		t.Fatalf("GetEntityByID: %v", err)
	}
	if got.CreatedAtTick != 100 {
		t.Errorf("created_at_tick: got %d, want 100", got.CreatedAtTick)
	}

	deathTick := int64(200)
	soft, err := q.SoftDeleteEntity(ctx, sqlc.SoftDeleteEntityParams{
		ID:              id,
		DestroyedAtTick: &deathTick,
	})
	if err != nil {
		t.Fatalf("SoftDeleteEntity: %v", err)
	}
	if soft.DestroyedAtTick == nil || *soft.DestroyedAtTick != deathTick {
		t.Errorf("destroyed_at_tick after soft delete: got %v, want %d", soft.DestroyedAtTick, deathTick)
	}

	// Soft-deleted row is still visible by ID — soft-delete preserves
	// the row for the audit log.
	if _, err := q.GetEntityByID(ctx, id); err != nil {
		t.Errorf("GetEntityByID after soft-delete: got %v, want row still readable", err)
	}

	// Second soft-delete is a no-op because of the partial WHERE.
	if _, err := q.SoftDeleteEntity(ctx, sqlc.SoftDeleteEntityParams{
		ID:              id,
		DestroyedAtTick: &deathTick,
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("idempotent soft-delete: got %v, want pgx.ErrNoRows", err)
	}
}

func TestEntityTypeCheckConstraint(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()

	if _, err := q.CreateEntity(ctx, sqlc.CreateEntityParams{
		ID:            newEntityID(t),
		SeasonID:      1,
		EntityType:    "dragon", // not one of the six allowed types
		CreatedAtTick: 1,
	}); err == nil {
		t.Fatal("expected CHECK violation for bogus entity_type")
	}
}

func TestListEntitiesByTypeInSeason(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()

	// Two characters + one NPC + one soft-deleted character.
	live1 := newEntityID(t)
	live2 := newEntityID(t)
	npc := newEntityID(t)
	dead := newEntityID(t)

	for _, p := range []sqlc.CreateEntityParams{
		{ID: live1, SeasonID: 1, EntityType: "character", CreatedAtTick: 1},
		{ID: live2, SeasonID: 1, EntityType: "character", CreatedAtTick: 2},
		{ID: npc, SeasonID: 1, EntityType: "npc", CreatedAtTick: 3},
		{ID: dead, SeasonID: 1, EntityType: "character", CreatedAtTick: 4},
	} {
		if _, err := q.CreateEntity(ctx, p); err != nil {
			t.Fatalf("CreateEntity(%s): %v", p.EntityType, err)
		}
	}

	destroyed := int64(5)
	if _, err := q.SoftDeleteEntity(ctx, sqlc.SoftDeleteEntityParams{
		ID:              dead,
		DestroyedAtTick: &destroyed,
	}); err != nil {
		t.Fatalf("SoftDeleteEntity: %v", err)
	}

	chars, err := q.ListEntitiesByTypeInSeason(ctx, sqlc.ListEntitiesByTypeInSeasonParams{
		SeasonID:   1,
		EntityType: "character",
	})
	if err != nil {
		t.Fatalf("ListEntitiesByTypeInSeason: %v", err)
	}
	if len(chars) != 2 {
		t.Errorf("live characters: got %d, want 2", len(chars))
	}
	for _, c := range chars {
		if c.ID == dead {
			t.Error("soft-deleted character should not appear in list")
		}
	}
}
