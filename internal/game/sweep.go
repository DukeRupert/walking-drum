package game

import (
	"context"
	"fmt"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
)

// SweepConfig controls the soft-delete sweep. Disabled by default per
// DESIGN.md §6.6 — the job is wired but gated until Layer 6 (audit
// log) lands and we can verify no late references exist.
type SweepConfig struct {
	// Enabled gates the whole job. False means SweepDestroyedEntities
	// is a no-op (returns 0, nil) even if everything else is set.
	Enabled bool

	// RetentionTicks is the grace period between soft-delete and hard-
	// delete. An entity becomes sweep-eligible when its
	// destroyed_at_tick is older than (currentTick - RetentionTicks).
	// Caller picks the window; the data layer doesn't know what a tick
	// means in wall-clock terms.
	RetentionTicks int64
}

// Querier is the subset of *sqlc.Queries the sweep needs. Narrowed so
// callers can pass either a pool-backed or tx-backed instance.
type Querier interface {
	SweepDestroyedEntities(ctx context.Context, destroyedAtTick *int64) (int64, error)
}

// SweepDestroyedEntities hard-deletes entities whose destroyed_at_tick
// is older than (currentTick - cfg.RetentionTicks). Returns the number
// of rows deleted. Cascading FKs handle dependent rows
// (entity_positions, components).
//
// No-ops (returns 0, nil) when cfg.Enabled is false — the production
// default until §6.6's preconditions are met.
func SweepDestroyedEntities(ctx context.Context, q Querier, currentTick int64, cfg SweepConfig) (int64, error) {
	if !cfg.Enabled {
		return 0, nil
	}
	if cfg.RetentionTicks < 0 {
		return 0, fmt.Errorf("sweep: negative retention (%d)", cfg.RetentionTicks)
	}
	cutoff := currentTick - cfg.RetentionTicks
	n, err := q.SweepDestroyedEntities(ctx, &cutoff)
	if err != nil {
		return 0, fmt.Errorf("sweep destroyed entities: %w", err)
	}
	return n, nil
}

// Compile-time check that *sqlc.Queries satisfies the Querier shape we
// need. If sqlc ever regenerates with a different signature, this will
// fail to build instead of silently breaking the sweep at runtime.
var _ Querier = (*sqlc.Queries)(nil)
