package game_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/game"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

func TestSweep_DisabledByDefault(t *testing.T) {
	_, tx := testdb.WithTx(t)
	ctx := context.Background()
	q := sqlc.New(tx)

	// Create + soft-delete an entity. With the sweep disabled, nothing
	// should disappear no matter how stale the row is.
	id, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1,
		Type:     game.EntityWorldObject,
		Tick:     1,
	})
	if err != nil {
		t.Fatalf("CreateEntity: %v", err)
	}
	destroyed := int64(2)
	if _, err := q.SoftDeleteEntity(ctx, sqlc.SoftDeleteEntityParams{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		DestroyedAtTick: &destroyed,
	}); err != nil {
		t.Fatalf("SoftDeleteEntity: %v", err)
	}

	n, err := game.SweepDestroyedEntities(ctx, q, 1_000_000, game.SweepConfig{
		Enabled:        false,
		RetentionTicks: 0,
	})
	if err != nil {
		t.Fatalf("SweepDestroyedEntities: %v", err)
	}
	if n != 0 {
		t.Errorf("disabled sweep deleted %d rows, want 0", n)
	}
	if _, err := q.GetEntityByID(ctx, pgtype.UUID{Bytes: id, Valid: true}); err != nil {
		t.Errorf("soft-deleted entity should still exist: %v", err)
	}
}

func TestSweep_RespectsRetentionWindow(t *testing.T) {
	_, tx := testdb.WithTx(t)
	ctx := context.Background()
	q := sqlc.New(tx)

	stale, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1, Type: game.EntityWorldObject, Tick: 1,
	})
	if err != nil {
		t.Fatalf("CreateEntity stale: %v", err)
	}
	fresh, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1, Type: game.EntityWorldObject, Tick: 1,
	})
	if err != nil {
		t.Fatalf("CreateEntity fresh: %v", err)
	}
	live, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1, Type: game.EntityWorldObject, Tick: 1,
	})
	if err != nil {
		t.Fatalf("CreateEntity live: %v", err)
	}

	// stale: destroyed long ago. fresh: destroyed inside the window.
	// live: not destroyed at all.
	staleTick := int64(10)
	freshTick := int64(95)
	for _, p := range []sqlc.SoftDeleteEntityParams{
		{ID: pgtype.UUID{Bytes: stale, Valid: true}, DestroyedAtTick: &staleTick},
		{ID: pgtype.UUID{Bytes: fresh, Valid: true}, DestroyedAtTick: &freshTick},
	} {
		if _, err := q.SoftDeleteEntity(ctx, p); err != nil {
			t.Fatalf("SoftDeleteEntity: %v", err)
		}
	}

	// currentTick=100, retention=20 → cutoff=80; only `stale` is older.
	n, err := game.SweepDestroyedEntities(ctx, q, 100, game.SweepConfig{
		Enabled:        true,
		RetentionTicks: 20,
	})
	if err != nil {
		t.Fatalf("SweepDestroyedEntities: %v", err)
	}
	if n != 1 {
		t.Errorf("swept rows: got %d, want 1", n)
	}

	if _, err := q.GetEntityByID(ctx, pgtype.UUID{Bytes: stale, Valid: true}); err == nil {
		t.Error("stale entity should be hard-deleted")
	}
	if _, err := q.GetEntityByID(ctx, pgtype.UUID{Bytes: fresh, Valid: true}); err != nil {
		t.Errorf("fresh (still in retention window) should remain: %v", err)
	}
	if _, err := q.GetEntityByID(ctx, pgtype.UUID{Bytes: live, Valid: true}); err != nil {
		t.Errorf("live entity should remain: %v", err)
	}
}

func TestSweep_CascadesToPositionAndComponents(t *testing.T) {
	_, tx := testdb.WithTx(t)
	ctx := context.Background()
	q := sqlc.New(tx)

	id, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1,
		Type:     game.EntityCharacter,
		Tick:     1,
		Position: &game.PositionSpec{RegionID: 1, X: 0, Y: 0},
		InitialComponents: []game.Component{
			game.Hidden{},
		},
	})
	if err != nil {
		t.Fatalf("CreateEntity: %v", err)
	}
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	destroyed := int64(2)
	if _, err := q.SoftDeleteEntity(ctx, sqlc.SoftDeleteEntityParams{
		ID:              pgID,
		DestroyedAtTick: &destroyed,
	}); err != nil {
		t.Fatalf("SoftDeleteEntity: %v", err)
	}

	n, err := game.SweepDestroyedEntities(ctx, q, 100, game.SweepConfig{
		Enabled:        true,
		RetentionTicks: 0,
	})
	if err != nil {
		t.Fatalf("SweepDestroyedEntities: %v", err)
	}
	if n != 1 {
		t.Errorf("swept rows: got %d, want 1", n)
	}

	if _, err := q.GetEntityPosition(ctx, pgID); err == nil {
		t.Error("position row should be cascade-deleted with the entity")
	}
	if _, err := q.GetComponent(ctx, sqlc.GetComponentParams{
		EntityID:      pgID,
		ComponentType: game.ComponentHidden,
	}); err == nil {
		t.Error("component row should be cascade-deleted with the entity")
	}
}
