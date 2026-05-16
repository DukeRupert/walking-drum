# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project status

Walking Drum is a server-authoritative multiplayer roguelike (Go backend, Postgres, eventual Svelte/WebSocket frontend). The repo is in early skeleton phase — Phase 0 of `TODO.md`. There is no game code yet; current work is wiring up the data layer.

**Always read `docs/DESIGN.md` before making non-trivial decisions.** It is the source of truth for architecture and schema. `TODO.md` tracks the phase-by-phase implementation sequence and is kept in sync with the design doc's "commit sequence" sections (§5.8, §6.6, §7.4). When design and code disagree, update one to match — don't silently diverge.

## Commands

```sh
# Start Postgres (requires POSTGRES_PASSWORD in the shell or .env)
docker compose up -d

# Run the smoke-test binary (loads .env, connects to DB)
go run ./cmd

# Tests
go test ./...
go test ./internal/envfile -run TestParse   # single test

# Tidy / build
go mod tidy
go build ./...
```

`goose` and `sqlc` are listed in Phase 0 but not yet installed/configured — when adding them, follow the design doc choices (`goose` for migrations, `sqlc` for query generation, generated code under `internal/db/` or similar).

`.env` is loaded by `cmd/main.go` via `internal/envfile`. Shell-set vars always win over `.env`. Required key: `DATABASE_URL`. `docker-compose.yml` also reads `POSTGRES_USER`/`POSTGRES_PASSWORD`/`POSTGRES_DB` from the environment.

## Architecture: the load-bearing ideas

These concepts thread through the whole codebase. Skim the design doc sections cited; they're short and they explain *why*.

- **Server-authoritative actor model** (§2.3): one Go binary holds all game state in memory. Per-region goroutines own mutable state; per-player session goroutines own WebSockets. Postgres is the durable store, not the runtime model. We are deliberately not introducing Redis, sharding, or multiple processes.

- **ECS-hybrid, not full ECS** (§2.4, §6.1, §7.1): every game object is an entity with a UUIDv7 ID. *Defining* state (HP on a character, energy on an actor) lives in dedicated normalized tables; *modifying* state (`OnFire`, `Stunned`) lives in the generic `components(entity_id, component_type, state JSONB)` table. Promote a component to its own table only when §4.3's criteria are met — and it's acceptable to promote on day one when those criteria are met immediately (`actors` is the precedent). Behavior is pure functions over world state, not methods on entities.

- **Energy-based time** (§2.5): game time is measured in monotonic ticks per season; wall-clock tick rate is a deploy-time knob, not a gameplay one. Actors regenerate energy; actions cost energy. Action costs are data (a future `action_costs` table), not constants in code. Anything that acts over time (player, NPC, projectile, fire patch) is an actor and goes through the same scheduler.

- **Six entity types, fixed** (§6.2): `character`, `npc`, `item`, `corpse`, `projectile`, `world_object`. Enforced via `CHECK` constraint, not enum (enums migrate painfully). Finer distinctions live in components and templates, not new entity types. `entity_type` is immutable — character→corpse is destroy+create, two IDs.

- **Seasons are the wipe boundary** (§3.5): the world resets every 3 months. Most tables carry `season_id`; everything except accounts, cosmetics, achievements, and leaderboards is wiped. Schema migrations that would be painful mid-season are deliberately timed to wipes. This shapes table design — e.g. audit log partitions by season for cheap drop-at-wipe.

- **Layered schema, committed in pieces** (§5–§8): the schema is structured as six layers (frame → entities/components → characters/items → world → trades → audit). Each layer is its own commit boundary. Cross-layer foreign keys are deliberately deferred — e.g. `entity_positions.region_id` is plain `INT` in Layer 2 and gets its FK added in Layer 4. Respect these layer seams when writing migrations.

## Code conventions specific to this repo

- **Module path:** `github.com/dukerupert/walking-drum`. Imports use the full path.
- **`internal/` is the home for non-public code.** `cmd/` holds the binary entrypoint(s).
- **DB access goes through `*pgxpool.Pool`** from `internal/db.Connect`. Pool config (limits, statement timeout) is set there — don't recreate it elsewhere.
- **UUIDv7 for entity IDs**, generated in Go (not Postgres). Choose a UUIDv7-supporting lib when entities land.
- **No raw tokens in the DB.** Session tokens, password reset tokens, etc. are stored hashed (§3.8, §5.5). The `token_hash` column pattern repeats — follow it.
- **Append-only tables (`moderation_actions`, audit log) are never updated.** Reversing an action means inserting a new row.
