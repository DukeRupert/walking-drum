package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
)

// TxBeginner is satisfied by both *pgxpool.Pool and pgx.Tx. The helper
// uses whichever it's handed: production code passes the pool and gets
// a real transaction; integration tests pass their already-rolling tx
// and get a nested savepoint, so the helper's atomicity is exercised
// without leaking rows across tests.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// PositionSpec is the optional position payload for CreateEntity. The
// caller leaves UpdatedAtTick zero and lets the helper fill it from
// the entity's CreatedAtTick — there's no meaningful "this position
// is older than the entity" case at creation time.
type PositionSpec struct {
	RegionID int32
	X, Y     int32
}

// CreateEntityInput is the one-stop payload for the entity-creation
// helper. Position and InitialComponents are optional; everything else
// is required because the underlying `entities` row needs it.
type CreateEntityInput struct {
	SeasonID          int32
	Type              EntityType
	Tick              int64
	Position          *PositionSpec
	InitialComponents []Component
}

// CreateEntity inserts an `entities` row plus the optional position and
// component rows in a single transaction. Returns the freshly generated
// UUIDv7 so the caller has a handle for follow-up work.
//
// This is the primary entity-creation API; Layer 3+ helpers
// (character-creation, NPC spawning, etc.) compose on top of it rather
// than rolling their own transactions.
func CreateEntity(ctx context.Context, tb TxBeginner, in CreateEntityInput) (uuid.UUID, error) {
	if !in.Type.Valid() {
		return uuid.Nil, fmt.Errorf("create entity: invalid type %q", in.Type)
	}
	id, err := NewEntityID()
	if err != nil {
		return uuid.Nil, err
	}
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	tx, err := tb.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin: %w", err)
	}
	// Rollback is a no-op if Commit already succeeded; harmless on the
	// happy path, essential on every error return below.
	defer func() { _ = tx.Rollback(ctx) }()

	q := sqlc.New(tx)

	if _, err := q.CreateEntity(ctx, sqlc.CreateEntityParams{
		ID:            pgID,
		SeasonID:      in.SeasonID,
		EntityType:    string(in.Type),
		CreatedAtTick: in.Tick,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("insert entity: %w", err)
	}

	if in.Position != nil {
		if _, err := q.SetEntityPosition(ctx, sqlc.SetEntityPositionParams{
			EntityID:      pgID,
			RegionID:      in.Position.RegionID,
			X:             in.Position.X,
			Y:             in.Position.Y,
			UpdatedAtTick: in.Tick,
		}); err != nil {
			return uuid.Nil, fmt.Errorf("insert position: %w", err)
		}
	}

	for _, c := range in.InitialComponents {
		if c == nil {
			return uuid.Nil, errors.New("create entity: nil component in InitialComponents")
		}
		raw, err := EncodeComponent(c)
		if err != nil {
			return uuid.Nil, err
		}
		if _, err := q.SetComponent(ctx, sqlc.SetComponentParams{
			EntityID:      pgID,
			ComponentType: c.ComponentType(),
			State:         raw,
			CreatedAtTick: in.Tick,
			UpdatedAtTick: in.Tick,
		}); err != nil {
			return uuid.Nil, fmt.Errorf("insert component %s: %w", c.ComponentType(), err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit: %w", err)
	}
	return id, nil
}
