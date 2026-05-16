// Package game holds the in-process model types for the world: entities,
// components, positions, and the helpers that build them on top of the
// data layer. It is deliberately Postgres-aware (it leans on the sqlc
// queries) but it sits *above* sqlc — callers should reach for these
// types rather than the generated row structs whenever they care about
// game semantics.
package game

import (
	"fmt"

	"github.com/google/uuid"
)

// EntityType is the six-value discriminator from DESIGN.md §6.2. The
// CHECK constraint on entities.entity_type is the source of truth;
// these constants exist so call sites don't sprinkle string literals.
type EntityType string

const (
	EntityCharacter   EntityType = "character"
	EntityNPC         EntityType = "npc"
	EntityItem        EntityType = "item"
	EntityCorpse      EntityType = "corpse"
	EntityProjectile  EntityType = "projectile"
	EntityWorldObject EntityType = "world_object"
)

// Valid reports whether t is one of the six allowed entity types. The
// DB will reject anything else via CHECK; this is a cheap early
// guard so we don't pay a round trip to find that out.
func (t EntityType) Valid() bool {
	switch t {
	case EntityCharacter, EntityNPC, EntityItem, EntityCorpse, EntityProjectile, EntityWorldObject:
		return true
	}
	return false
}

// Entity is the in-memory shape of one row of `entities`. DestroyedAtTick
// is zero when the entity is live; the soft-delete sweep reads this.
type Entity struct {
	ID               uuid.UUID
	SeasonID         int32
	Type             EntityType
	CreatedAtTick    int64
	DestroyedAtTick  int64 // 0 = live
	IsDestroyed      bool
}

// Position is the in-memory shape of one row of `entity_positions`.
// Sparse: not every entity has one (abstract / container-held entities).
type Position struct {
	EntityID       uuid.UUID
	RegionID       int32
	X, Y           int32
	UpdatedAtTick  int64
}

// Component is the marker interface every typed component struct
// satisfies. The Type() return value is what lands in
// components.component_type; the struct itself is JSON-serialized
// into components.state.
type Component interface {
	ComponentType() string
}

// NewEntityID returns a fresh UUIDv7. Centralized so we have one place
// to swap libraries if we ever outgrow google/uuid's implementation.
func NewEntityID() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("new uuidv7: %w", err)
	}
	return id, nil
}
