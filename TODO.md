## Phase 0 — Project Skeleton

Bones of the project. Get this working before touching any game code.

- [x] Go module init — module path, `go.mod`, basic directory structure (`cmd/`, `internal/`, `migrations/`)
- [x] PostgreSQL running locally via Docker Compose — single Postgres service, named volume for persistence, exposed port, connection string in environment
- [x] `pgx` and `pgxpool` wired in — a `db.Connect()` function returning `*pgxpool.Pool`, reads connection string from env, sensible pool config
- [x] `goose` installed and configured — `migrations/` directory recognized, `goose status` works, a no-op initial migration verifies up/down cycle
- [x] `sqlc` installed and configured — `sqlc.yaml` pointing at `migrations/` for schema and `queries/` directory for query files, generated code lands in `internal/db/` or similar, empty initial generation works
- [x] Smoke test — a `main.go` that connects to the DB, runs migrations, prints "ok," exits

**Done when:** smoke test runs cleanly end-to-end.

---

## Phase 1 — Layer 1: The Frame

Per §5.8 of the design doc.

- [x] Migration: `citext` extension (trivial, but proves migrations work end-to-end)
- [x] Migration: `accounts` + `account_flags` tables, including the partial index on `status`
- [x] `sqlc` queries for accounts — create account, look up by email, look up by id, update status, soft-delete
- [x] Migration: `seasons` + `season_participation` tables, including the unique partial index for "one active season at a time" and a seed migration creating season 1 with placeholder dates and seed
- [x] `sqlc` queries for seasons — get active season, get season by id, advance season status (`upcoming` → `active` → `ended`)
- [x] Migration: `sessions` table, including both partial indexes
- [x] `sqlc` queries for sessions — create session, look up by token hash, revoke session, list active sessions for an account
- [x] Migration: `moderation_actions` table, including both indexes
- [x] `sqlc` queries for moderation — append action, list actions for an account, find currently-active suspensions/bans
- [x] Auth helper layer (not yet HTTP) — `bcrypt` wrapper for password hashing, token generator (random bytes → hash for storage), session-creation function taking an account and producing a session row

**Done when:** a small Go program can create an account, create a session, validate the session token, and revoke it. ✅

---

## Phase 2 — Layer 2: Entities and Components

Per §6.6 of the design doc.

- [x] Migration: `entities` table with CHECK constraint listing the six entity types and both indexes (no FKs to it from other tables yet)
- [x] Migration: `entity_positions` table — `region_id` as plain `INT NOT NULL`, no FK yet, both indexes
- [x] Migration: `components` table with the reverse-direction index `(component_type, entity_id)`
- [x] Go types in `internal/game/` (or similar) — `Entity` struct, `Position` struct, `Component` interface or marker type, UUIDv7 generation helper (use `github.com/google/uuid` v1.6+ or a dedicated UUIDv7 lib)
- [x] First concrete component as smoke test — something cheap and marker-shaped (`Hidden{}` or `Persistent{}`); define the typed Go struct, write the JSONB serialization round-trip test
- [x] `sqlc` queries for entities — create entity (just the row), look up by ID, soft-delete (set `destroyed_at_tick`), list by `(season_id, entity_type)` filter
- [x] `sqlc` queries for entity_positions — set position (upsert), get position by entity_id, delete position by entity_id, query entities at `(region_id, x, y)`, query entities in region
- [x] `sqlc` queries for components — set component (upsert keyed on `(entity_id, component_type)`), get component, delete component, iterate entities with a given component type (joined to `entities.destroyed_at_tick IS NULL`)
- [x] Transactional entity-creation helper in Go — given a season, type, optional position, and optional initial components, creates the entity row + position row + component rows in a single transaction
- [ ] Sweep job stub — hard-deletes entities where `destroyed_at_tick < (current_tick - retention_window)`, wired but gated by config flag, disabled by default until Layer 6 is real

**Done when:** a test creates a fictional entity with a position and a marker component, looks it up, updates the component, soft-deletes the entity, and confirms the row is still there but the position row is gone.

---

## Phase 3 — Layer 3 Piece 1: Characters and Actors

Per §7.4 of the design doc.

- [ ] Migration: `characters` table with all three indexes (unique name index, active-account partial index, hp-zero partial index)
- [ ] Migration: `actors` table with the partial scheduling index on `next_ready_tick`
- [ ] Go types — `Character` and `Actor` structs in the game package
- [ ] `sqlc` queries for characters — create character (just the row), look up living character for an account in a given season, apply damage (UPDATE hp = hp - $1), mark character dead (set `died_at_tick`), query dead-but-not-yet-processed characters for the death sweep
- [ ] `sqlc` queries for actors — insert actor row, update `next_ready_tick` (the `MarkActorReady` primitive), scheduler query (entities with `next_ready_tick <= $current_tick`, ordered), deduct energy and recompute `next_ready_tick`
- [ ] Character-creation transaction helper — builds on Phase 2's entity-creation helper; creates `entities` row + `entity_positions` row + `actors` row + `characters` row in one transaction sharing the same entity_id
- [ ] Death transaction helper — sets `died_at_tick` on `characters`, sets `destroyed_at_tick` on `entities`, deletes the `entity_positions` row, all in one transaction (does NOT yet handle the corpse-and-inventory flow — that's a future Layer 3 piece)
- [ ] `MarkActorReady` helper function in Go — single chokepoint for re-enabling actor scheduling per §7.3, wraps the underlying `sqlc` query

**Done when:** a test creates a character, applies damage until HP reaches zero, runs the death sweep, and confirms the character is marked dead, the entity is soft-deleted, the position row is gone, and the actor row is still there.

---

## Testing Cadence

For each `sqlc` query group, write at least one integration test that exercises it against a real Postgres (test container or dedicated test DB). Catching schema-vs-query mismatches at this level is much cheaper than catching them at system-test level.

---

## Explicitly Not in This List

These belong to later phases or different conversations:

- HTTP layer, auth handlers, signup/login flows — data layer is independently testable without HTTP
- WebSocket layer, region goroutines, scheduler loop — can't be fully exercised until there are NPCs and a tick loop (later Layer 3 pieces)
- Admin tooling — §3.6 calls it out as v1, but it sits on top of the data layer
- Backups, monitoring, deployment — local dev setup is enough for now
- Layer 3 remaining pieces (items, Location, corpses, action_costs, projectiles, templates) — continue schema work in alternation with implementation

---

## Natural Stopping Points

Three good places to pause:

1. End of Phase 0 — project skeleton runnable
2. End of Phase 1 — auth round-trip works
3. End of Phase 2 — entity round-trip works

Phase 3 is the first one that builds on game-specific concepts; decision friction inside Phase 3 is the signal to pause implementation and resume design.FLICT UPDATE — upsert semantics).
Get position by entity_id.
Delete position by entity_id.
Query entities at (region_id, x, y).
Query entities in region.


sqlc queries for components.

Set component on entity (upsert keyed on (entity_id, component_type)).
Get component by (entity_id, component_type).
Delete component.
Iterate entities with a given component type (joined to entities.destroyed_at_tick IS NULL).


Transactional entity-creation helper in Go. A function that, given a season, type, optional position, and optional initial components, creates the entity row + position row + component rows in a single Postgres transaction. This is the primary entity-creation API for everything downstream.
Sweep job stub. A function that hard-deletes entities where destroyed_at_tick < (current_tick - retention_window). Wired but gated by a config flag, disabled by default until Layer 6 is real.

Phase 2 is done when you can write a test that creates a fictional entity with a position and a marker component, looks it up, updates the component, soft-deletes the entity, and confirms the row is still there but the position row is gone.
Phase 3 — Layer 3 Piece 1: Characters and Actors
Per §7.4:

Migration: characters table. With all three indexes (the unique name index, the active-account partial index, the hp-zero partial index).
Migration: actors table. With the partial scheduling index on next_ready_tick.
Go types. Character and Actor structs in the game package.
sqlc queries for characters.

Create character (just the characters row — the wrapping transaction comes in step 32).
Look up living character for an account in a given season.
Apply damage (UPDATE hp = hp - $1).
Mark character dead (set died_at_tick).
Query dead-but-not-yet-processed characters for the death sweep.


sqlc queries for actors.

Insert actor row.
Update next_ready_tick (the MarkActorReady primitive).
Scheduler query — entities with next_ready_tick <= $current_tick, ordered.
Deduct energy and recompute next_ready_tick.


Character-creation transaction helper. Builds on the Phase 2 entity-creation helper: creates entities row + entity_positions row + actors row + characters row, all in one transaction, all sharing the same entity_id. This is the function the eventual "spawn character at village" code path will call.
Death transaction helper. Sets died_at_tick on characters, sets destroyed_at_tick on entities, deletes the entity_positions row, all in one transaction. Doesn't yet handle the corpse-and-inventory flow — that's a future piece of Layer 3.
MarkActorReady helper function in Go. The single chokepoint for re-enabling actor scheduling, as called out in §7.3. Wraps the underlying sqlc query.

Phase 3 is done when you can write a test that creates a character, applies damage to it until HP reaches zero, runs the death sweep, and confirms the character is marked dead, the entity is soft-deleted, the position row is gone, and the actor row is still there (cascade-deletable when the entity is hard-deleted later).