package game_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/game"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

// TestPhase2DoneWhen is the round-trip described in TODO.md's
// "Done when" line for Phase 2: create a fictional entity with a
// position and a marker component, look it up, update the component,
// soft-delete the entity, and confirm the row is still there but the
// position row is gone.
func TestPhase2DoneWhen(t *testing.T) {
	_, tx := testdb.WithTx(t)
	ctx := context.Background()
	q := sqlc.New(tx)

	// 1. Create entity with position + Hidden component, all in one tx.
	id, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1,
		Type:     game.EntityWorldObject,
		Tick:     100,
		Position: &game.PositionSpec{RegionID: 4, X: 11, Y: 22},
		InitialComponents: []game.Component{
			game.Hidden{},
		},
	})
	if err != nil {
		t.Fatalf("CreateEntity: %v", err)
	}
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	// 2. Look it up — entity, position, and component all readable.
	if _, err := q.GetEntityByID(ctx, pgID); err != nil {
		t.Fatalf("GetEntityByID: %v", err)
	}
	pos, err := q.GetEntityPosition(ctx, pgID)
	if err != nil {
		t.Fatalf("GetEntityPosition: %v", err)
	}
	if pos.RegionID != 4 || pos.X != 11 || pos.Y != 22 {
		t.Errorf("position: got (%d, %d, %d), want (4, 11, 22)", pos.RegionID, pos.X, pos.Y)
	}
	if _, err := q.GetComponent(ctx, sqlc.GetComponentParams{
		EntityID:      pgID,
		ComponentType: game.ComponentHidden,
	}); err != nil {
		t.Fatalf("GetComponent: %v", err)
	}

	// 3. Update the component. SetComponent is an upsert; the
	// created_at_tick should survive while updated_at_tick advances.
	updated, err := q.SetComponent(ctx, sqlc.SetComponentParams{
		EntityID:      pgID,
		ComponentType: game.ComponentHidden,
		State:         []byte(`{"reason":"gm"}`),
		CreatedAtTick: 999, // ignored on conflict
		UpdatedAtTick: 200,
	})
	if err != nil {
		t.Fatalf("SetComponent (update): %v", err)
	}
	if updated.CreatedAtTick != 100 {
		t.Errorf("created_at_tick after update: got %d, want 100", updated.CreatedAtTick)
	}
	if updated.UpdatedAtTick != 200 {
		t.Errorf("updated_at_tick after update: got %d, want 200", updated.UpdatedAtTick)
	}

	// 4. Soft-delete the entity, and (per DESIGN.md §6.3) drop the
	// position row. The entity row itself stays — soft-deleted rows
	// are visible until the periodic hard-delete sweep.
	destroyed := int64(300)
	if _, err := q.SoftDeleteEntity(ctx, sqlc.SoftDeleteEntityParams{
		ID:              pgID,
		DestroyedAtTick: &destroyed,
	}); err != nil {
		t.Fatalf("SoftDeleteEntity: %v", err)
	}
	if err := q.DeleteEntityPosition(ctx, pgID); err != nil {
		t.Fatalf("DeleteEntityPosition: %v", err)
	}

	// 5. Confirm: entity row still there (soft-deleted), position gone.
	ent, err := q.GetEntityByID(ctx, pgID)
	if err != nil {
		t.Fatalf("GetEntityByID after soft-delete: %v", err)
	}
	if ent.DestroyedAtTick == nil || *ent.DestroyedAtTick != destroyed {
		t.Errorf("destroyed_at_tick: got %v, want %d", ent.DestroyedAtTick, destroyed)
	}
	if _, err := q.GetEntityPosition(ctx, pgID); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("GetEntityPosition after delete: got %v, want pgx.ErrNoRows", err)
	}
}
