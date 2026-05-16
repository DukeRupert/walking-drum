package sqlc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

func TestComponentUpsert(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	id := makeLiveEntity(t, ctx, q, 1)

	first, err := q.SetComponent(ctx, sqlc.SetComponentParams{
		EntityID:      id,
		ComponentType: "hidden",
		State:         []byte("{}"),
		CreatedAtTick: 10,
		UpdatedAtTick: 10,
	})
	if err != nil {
		t.Fatalf("SetComponent (insert): %v", err)
	}
	if first.CreatedAtTick != 10 || first.UpdatedAtTick != 10 {
		t.Errorf("ticks after insert: created=%d updated=%d, want both 10", first.CreatedAtTick, first.UpdatedAtTick)
	}

	// Upsert: same key, new state, later tick. created_at_tick should
	// stick to the original; updated_at_tick should advance.
	second, err := q.SetComponent(ctx, sqlc.SetComponentParams{
		EntityID:      id,
		ComponentType: "hidden",
		State:         []byte(`{"updated":true}`),
		CreatedAtTick: 999, // ignored by the ON CONFLICT path
		UpdatedAtTick: 20,
	})
	if err != nil {
		t.Fatalf("SetComponent (update): %v", err)
	}
	if second.CreatedAtTick != 10 {
		t.Errorf("created_at_tick after upsert: got %d, want 10 (stuck)", second.CreatedAtTick)
	}
	if second.UpdatedAtTick != 20 {
		t.Errorf("updated_at_tick after upsert: got %d, want 20", second.UpdatedAtTick)
	}
	if string(second.State) != `{"updated": true}` && string(second.State) != `{"updated":true}` {
		t.Errorf("state after upsert: got %q, want JSON {\"updated\":true}", string(second.State))
	}
}

func TestComponentGetAndDelete(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()
	id := makeLiveEntity(t, ctx, q, 1)

	if _, err := q.SetComponent(ctx, sqlc.SetComponentParams{
		EntityID:      id,
		ComponentType: "hidden",
		State:         []byte("{}"),
		CreatedAtTick: 1,
		UpdatedAtTick: 1,
	}); err != nil {
		t.Fatalf("SetComponent: %v", err)
	}

	got, err := q.GetComponent(ctx, sqlc.GetComponentParams{
		EntityID:      id,
		ComponentType: "hidden",
	})
	if err != nil {
		t.Fatalf("GetComponent: %v", err)
	}
	if got.ComponentType != "hidden" {
		t.Errorf("got component_type %q, want hidden", got.ComponentType)
	}

	if err := q.DeleteComponent(ctx, sqlc.DeleteComponentParams{
		EntityID:      id,
		ComponentType: "hidden",
	}); err != nil {
		t.Fatalf("DeleteComponent: %v", err)
	}
	if _, err := q.GetComponent(ctx, sqlc.GetComponentParams{
		EntityID:      id,
		ComponentType: "hidden",
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("GetComponent after delete: got %v, want pgx.ErrNoRows", err)
	}
}

func TestListEntitiesWithComponentSkipsSoftDeleted(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()

	live := makeLiveEntity(t, ctx, q, 1)
	dead := makeLiveEntity(t, ctx, q, 2)
	other := makeLiveEntity(t, ctx, q, 3)

	// Two entities marked hidden, one persistent-only (separate type).
	for _, p := range []sqlc.SetComponentParams{
		{EntityID: live, ComponentType: "hidden", State: []byte("{}"), CreatedAtTick: 1, UpdatedAtTick: 1},
		{EntityID: dead, ComponentType: "hidden", State: []byte("{}"), CreatedAtTick: 1, UpdatedAtTick: 1},
		{EntityID: other, ComponentType: "persistent", State: []byte("{}"), CreatedAtTick: 1, UpdatedAtTick: 1},
	} {
		if _, err := q.SetComponent(ctx, p); err != nil {
			t.Fatalf("SetComponent: %v", err)
		}
	}

	// Soft-delete one of the hidden entities; the iterator must skip it.
	dt := int64(5)
	if _, err := q.SoftDeleteEntity(ctx, sqlc.SoftDeleteEntityParams{
		ID:              dead,
		DestroyedAtTick: &dt,
	}); err != nil {
		t.Fatalf("SoftDeleteEntity: %v", err)
	}

	rows, err := q.ListEntitiesWithComponent(ctx, "hidden")
	if err != nil {
		t.Fatalf("ListEntitiesWithComponent: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("hidden rows on live entities: got %d, want 1", len(rows))
	}
	if rows[0].EntityID != live {
		t.Error("ListEntitiesWithComponent returned the wrong entity")
	}
}
