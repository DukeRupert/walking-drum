package sqlc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

// makeLiveEntity inserts a minimal `character` entity in season 1 and
// returns its ID. Shared by the positions and components tests; both
// need an entity to point rows at.
func makeLiveEntity(t *testing.T, ctx context.Context, q *sqlc.Queries, tick int64) pgtype.UUID {
	t.Helper()
	id := newEntityID(t)
	if _, err := q.CreateEntity(ctx, sqlc.CreateEntityParams{
		ID:            id,
		SeasonID:      1,
		EntityType:    "character",
		CreatedAtTick: tick,
	}); err != nil {
		t.Fatalf("makeLiveEntity: %v", err)
	}
	return id
}

func TestEntityPositionUpsertAndGet(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	id := makeLiveEntity(t, ctx, q, 1)

	first, err := q.SetEntityPosition(ctx, sqlc.SetEntityPositionParams{
		EntityID:      id,
		RegionID:      7,
		X:             3,
		Y:             4,
		UpdatedAtTick: 10,
	})
	if err != nil {
		t.Fatalf("SetEntityPosition (insert): %v", err)
	}
	if first.X != 3 || first.Y != 4 || first.RegionID != 7 {
		t.Errorf("inserted position: got (region=%d, x=%d, y=%d), want (7, 3, 4)", first.RegionID, first.X, first.Y)
	}

	// Upsert: same entity, different coords.
	second, err := q.SetEntityPosition(ctx, sqlc.SetEntityPositionParams{
		EntityID:      id,
		RegionID:      7,
		X:             5,
		Y:             6,
		UpdatedAtTick: 11,
	})
	if err != nil {
		t.Fatalf("SetEntityPosition (update): %v", err)
	}
	if second.X != 5 || second.Y != 6 {
		t.Errorf("upserted position: got (x=%d, y=%d), want (5, 6)", second.X, second.Y)
	}
	if second.UpdatedAtTick != 11 {
		t.Errorf("updated_at_tick not bumped: got %d, want 11", second.UpdatedAtTick)
	}

	got, err := q.GetEntityPosition(ctx, id)
	if err != nil {
		t.Fatalf("GetEntityPosition: %v", err)
	}
	if got.X != 5 || got.Y != 6 {
		t.Errorf("GetEntityPosition: got (x=%d, y=%d), want (5, 6)", got.X, got.Y)
	}
}

func TestEntityPositionDelete(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	id := makeLiveEntity(t, ctx, q, 1)

	if _, err := q.SetEntityPosition(ctx, sqlc.SetEntityPositionParams{
		EntityID:      id,
		RegionID:      1,
		X:             0,
		Y:             0,
		UpdatedAtTick: 1,
	}); err != nil {
		t.Fatalf("SetEntityPosition: %v", err)
	}

	if err := q.DeleteEntityPosition(ctx, id); err != nil {
		t.Fatalf("DeleteEntityPosition: %v", err)
	}
	if _, err := q.GetEntityPosition(ctx, id); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("GetEntityPosition after delete: got %v, want pgx.ErrNoRows", err)
	}

	// Deleting again is a no-op (silently zero rows).
	if err := q.DeleteEntityPosition(ctx, id); err != nil {
		t.Errorf("DeleteEntityPosition (idempotent): %v", err)
	}
}

func TestEntityPositionSpatialQueries(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()

	// Place three entities: two at the same tile in region 1, one
	// at a different tile in the same region, one in another region.
	a := makeLiveEntity(t, ctx, q, 1)
	b := makeLiveEntity(t, ctx, q, 2)
	c := makeLiveEntity(t, ctx, q, 3)
	d := makeLiveEntity(t, ctx, q, 4)

	for _, p := range []sqlc.SetEntityPositionParams{
		{EntityID: a, RegionID: 1, X: 0, Y: 0, UpdatedAtTick: 1},
		{EntityID: b, RegionID: 1, X: 0, Y: 0, UpdatedAtTick: 1},
		{EntityID: c, RegionID: 1, X: 5, Y: 5, UpdatedAtTick: 1},
		{EntityID: d, RegionID: 2, X: 0, Y: 0, UpdatedAtTick: 1},
	} {
		if _, err := q.SetEntityPosition(ctx, p); err != nil {
			t.Fatalf("SetEntityPosition: %v", err)
		}
	}

	at00, err := q.GetEntitiesAtPosition(ctx, sqlc.GetEntitiesAtPositionParams{
		RegionID: 1,
		X:        0,
		Y:        0,
	})
	if err != nil {
		t.Fatalf("GetEntitiesAtPosition: %v", err)
	}
	if len(at00) != 2 {
		t.Errorf("entities at (1, 0, 0): got %d, want 2", len(at00))
	}

	region1, err := q.GetEntitiesInRegion(ctx, 1)
	if err != nil {
		t.Fatalf("GetEntitiesInRegion: %v", err)
	}
	if len(region1) != 3 {
		t.Errorf("entities in region 1: got %d, want 3", len(region1))
	}
}
