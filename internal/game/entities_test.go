package game_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/game"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

func TestCreateEntity_BareMinimum(t *testing.T) {
	_, tx := testdb.WithTx(t)
	ctx := context.Background()

	id, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1,
		Type:     game.EntityWorldObject,
		Tick:     42,
	})
	if err != nil {
		t.Fatalf("CreateEntity: %v", err)
	}

	// The helper opened a savepoint subtx and committed it; rows are
	// visible from the outer test tx.
	q := sqlc.New(tx)
	ent, err := q.GetEntityByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		t.Fatalf("GetEntityByID: %v", err)
	}
	if ent.EntityType != "world_object" {
		t.Errorf("entity_type: got %q, want world_object", ent.EntityType)
	}

	// No position requested, no position row.
	if _, err := q.GetEntityPosition(ctx, ent.ID); err == nil {
		t.Error("expected no position row, got one")
	}
}

func TestCreateEntity_WithPositionAndComponents(t *testing.T) {
	_, tx := testdb.WithTx(t)
	ctx := context.Background()

	id, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1,
		Type:     game.EntityCharacter,
		Tick:     100,
		Position: &game.PositionSpec{RegionID: 3, X: 10, Y: 20},
		InitialComponents: []game.Component{
			game.Hidden{},
		},
	})
	if err != nil {
		t.Fatalf("CreateEntity: %v", err)
	}

	q := sqlc.New(tx)
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	pos, err := q.GetEntityPosition(ctx, pgID)
	if err != nil {
		t.Fatalf("GetEntityPosition: %v", err)
	}
	if pos.RegionID != 3 || pos.X != 10 || pos.Y != 20 {
		t.Errorf("position: got (%d, %d, %d), want (3, 10, 20)", pos.RegionID, pos.X, pos.Y)
	}
	if pos.UpdatedAtTick != 100 {
		t.Errorf("updated_at_tick: got %d, want 100 (=entity tick)", pos.UpdatedAtTick)
	}

	comp, err := q.GetComponent(ctx, sqlc.GetComponentParams{
		EntityID:      pgID,
		ComponentType: game.ComponentHidden,
	})
	if err != nil {
		t.Fatalf("GetComponent: %v", err)
	}
	var h game.Hidden
	if err := game.DecodeComponent(comp.State, &h); err != nil {
		t.Fatalf("DecodeComponent: %v", err)
	}
}

func TestCreateEntity_RejectsBadType(t *testing.T) {
	_, tx := testdb.WithTx(t)
	ctx := context.Background()

	if _, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1,
		Type:     game.EntityType("dragon"),
		Tick:     1,
	}); err == nil {
		t.Fatal("expected error for invalid entity type")
	}
}

func TestCreateEntity_AtomicityOnComponentFailure(t *testing.T) {
	_, tx := testdb.WithTx(t)
	ctx := context.Background()

	// Trip the nil-component guard mid-flight. The entity and position
	// inserts succeeded first; rollback must drop them too.
	_, err := game.CreateEntity(ctx, tx, game.CreateEntityInput{
		SeasonID: 1,
		Type:     game.EntityCharacter,
		Tick:     1,
		Position: &game.PositionSpec{RegionID: 1, X: 0, Y: 0},
		InitialComponents: []game.Component{
			game.Hidden{},
			nil, // boom
		},
	})
	if err == nil {
		t.Fatal("expected error from nil component, got nil")
	}

	// No live characters in season 1 should exist as a side-effect of
	// the failed call.
	rows, err := sqlc.New(tx).ListEntitiesByTypeInSeason(ctx, sqlc.ListEntitiesByTypeInSeasonParams{
		SeasonID:   1,
		EntityType: "character",
	})
	if err != nil {
		t.Fatalf("ListEntitiesByTypeInSeason: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("entity rows after rolled-back create: got %d, want 0", len(rows))
	}
}
