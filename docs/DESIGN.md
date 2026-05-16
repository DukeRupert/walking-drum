# Multiplayer Roguelike — Design Document

**Working title:** Walking Drum
**Author:** Logan Williams
**Status:** Living document — updated as decisions are made
**Last updated:** May 15, 2026

---

## 1. Game Concept

A multiplayer roguelike in the vein of NetHack and Caves of Qud, delivered as a web application.

**Core loop:** All players start in a shared starting village. The world is shared — one persistent world state, one map, one economy. Players explore outward from the starting village. The world gets more lucrative and more dangerous the farther a player travels from the starting area.

**Hardcore permadeath:** On death, a player loses all character progress. Their inventory remains on their corpse at the location of death, recoverable by themselves or other players.

**Social model:** Trading and teaming up are allowed anywhere. **No player-vs-player combat.** The game is focused on exploration, survival, and cooperation against the world — not against other players.

**Seasons:** The world resets every 3 months. Wipes are scheduled events used both to refresh the game economy and to deploy major updates and schema migrations.

---

## 2. Architectural Decisions

### 2.1 Authority Model

**Server-authoritative.** All game state lives on the server. The client requests actions; the server validates and applies them. The client never directly mutates game state.

Rationale: multiplayer + hardcore permadeath + economic activity (trading) means cheating defense is a hard requirement. Server-authoritative is the foundation of that defense.

### 2.2 Tech Stack

**Backend:** Go, single binary, holding authoritative game state in-process.
**Frontend:** Svelte + TypeScript. Tile-based rendering (specific approach — DOM grid vs. canvas — deferred to the frontend layer discussion).
**Transport:** WebSocket for live game state. HTTP for auth, lobby, account management, admin tools.
**Database:** PostgreSQL.
**Hosting:** Single Hetzner VPS to start. One Go binary + one Postgres instance.

**Explicitly deferred:** Redis/Valkey, sharding, multiple game-server processes. Add only when measured need arises. Keep it simple until absolutely necessary.

### 2.3 Server Architecture (high level)

**Spatial partitioning by region.** The world is divided into regions (size TBD). Each region is owned by a goroutine that:

- Owns all mutable state in that region
- Receives player action messages via a channel
- Broadcasts state updates to subscribed players
- Persists its state to Postgres on a cadence (snapshot on quiet ticks, on player departure, on graceful shutdown)
- Can hibernate when no players are present

**Per-player session goroutine.** Each connected player has a goroutine handling their WebSocket. The session subscribes to the player's current region and transitions subscription on region change.

**Central hub** for matchmaking, region routing, and global concerns (chat, death feed, etc.).

This is the actor-model pattern, well-suited to Go's concurrency primitives. The architecture has the seams (region-based) to split across processes later if ever needed — but we don't plan to.

### 2.4 In-Memory Model: ECS-Flavored Hybrid

Conceptual model the runtime and persistence layer both follow:

- **Everything in the world is an entity with a stable ID** (UUIDv7).
- **Entities have two kinds of attributes:**
  - **Core attributes** — things the entity definitely has by virtue of what it is (character name, position, HP). Live in conventional normalized tables and well-typed Go structs.
  - **Components** — sparse, optional, extensible state (OnFire, Stunned, AIGoal). Live in a generic component table on disk and `map[EntityID]ComponentState` in memory.
- **Behavior lives in systems** — pure functions: `(world_state, action) -> new_world_state`. No methods on entity classes.
- **Content lives in data** — item templates, monster templates, region prefabs are database rows or loaded data files, not Go code.

**We are not adopting a full Go ECS framework (donburi, arche, etc.) for v1.** We're applying ECS-style discipline to a conventional Go architecture. The path to a full framework remains open if the game grows in that direction.

**Specific runtime discipline:**

- Entity IDs as the universal handle. Never pass pointers to game objects between systems.
- Optional components as sparse maps, not nullable fields.
- Systems as pure functions of world state (mutate in place underneath if needed, but think functionally).
- Prefabs / templates for content. Designers (eventually) create content as data, not code.

### 2.5 Time System

**Energy-based, server-driven.** Inspired by Cogmind and the Korostelev energy-scheduler pattern. Every actor (player character, NPC, projectile, fire patch, regen effect) has energy that regenerates over time; actions cost energy. When an actor has enough energy, they may act.

**Tick rate is a network-layer concern, not a gameplay one.** The server advances simulation on a fixed tick (e.g., 100ms or 250ms — tunable in deployment config). Changing the tick rate changes broadcast granularity and server cost, not game balance.

**Action costs are gameplay data.** "Move = 40 energy, attack = 200, fire a volley of 5 weapons = 375" — these are the gameplay dials. They live in a data table (Layer 3), seeded by migration, and overridable per-season via the season modifiers JSONB.

**Energy regen is per-entity** — a property of the `Actor` component (Layer 3). A potion of haste increases regen rate; carrying a heavy load decreases it.

**Game-time is measured in ticks**, monotonic per season, reset to 0 on wipe. A global `current_tick` counter lives in the server's central hub, persisted periodically. Wall-clock timestamps still exist for ops/audit purposes, but game-tick is the source of truth for "when did this happen in the world."

**Key insight:** energy is a game-currency, not a wall-clock measure. It doesn't matter whether 100 energy regen happens over 1 second or 2.5 seconds of wall time — what matters is that move costs 40 of it and attack costs 200. The tick rate is just how finely the server slices the regen.

---

## 3. Multiplayer Design Decisions

### 3.1 Multiplayer Flavor

**Shared-world persistent.** All players exist in one world. Spatial locality matters — players far apart don't need to know about each other in real time. Player density varies wildly between the village (crowded) and the frontier (sparse).

### 3.2 Idle / AFK Behavior

- **In safe zones:** idle players are kept alive indefinitely. The world cannot kill them.
- **In unsafe zones:** idle players are eventually killed.
  - **Passive death:** wandering monsters can find and kill them per normal world rules.
  - **Active death backstop:** after a long idle period (~30 minutes, exact value TBD) of zero input in an unsafe zone, the server kills the character with narrative flavor (starvation, exposure). Prevents the "log out in a forgotten corner" loophole.

### 3.3 Offline Player / Disconnection Behavior

- **Clean disconnect (client closes WebSocket cleanly):** character remains in the world per normal rules. No grace, no special handling. This prevents rage-quit-to-escape exploits.
- **Connection drop (network failure, timeout):** character becomes invulnerable for a grace period (~90s to 2min, exact value TBD).
  - During the grace period, the character cannot act, be acted upon, or be looted.
  - On reconnect within the grace window, the player is offered: "retreat to safety" or "resume where you were."
  - On timeout (no reconnection in the grace window), the character is auto-teleported to the starting safe zone with all inventory intact.
- **Exploit hardening:** repeated disconnects in a short window reduce the available grace period (suspicious behavior detection). Grace period is non-actionable — players cannot gain advantage from disconnecting mid-fight.
- **Mid-trade disconnection:** trades are cancelled gracefully, all items returned to original owners atomically.

### 3.4 Safe Zones

- **Binary** (safe / not safe) for v1, not graduated. Can be revisited later.
- **Static** — set at world generation, do not shift during a season.
- Modeled as a property on regions (or sub-regions).
- Effects: no world damage, idle players stay alive, hostile mob spawns disabled, item decay on corpses disabled, conveniences enabled (trade UI).

### 3.5 Seasons

- **Duration:** 3 months.
- **What persists across seasons:** accounts, login credentials, cosmetic unlocks, achievements, lifetime stats, leaderboards, hall of fame entries.
- **What is wiped:** characters, inventories, world state, corpses, learned skills, regions, generated content.
- **End-of-season process:** advance announcement (1–2 weeks), play closure at season end, leaderboard/hall-of-fame snapshot generation, world data truncation, new world generation from new seed, gates reopen.
- **The wipe is also a chance to apply schema migrations that would be painful mid-season.** Plan around this cadence.
- **The wipe operation must be testable on staging long before the first real wipe.**

### 3.6 Cheating Defense

**Tier 1 — foundational, built during the prototype:**

- Server-authoritative everything.
- Action validation on every input (distance, cooldown, weight, line-of-sight).
- Rate limiting per connection.
- Audit log for high-value events (deaths, trades, item creation, level-ups).

**Tier 2 — pre-launch:**

- Account creation friction (email verification, CAPTCHA at signup).
- Anomaly detection on action patterns.
- Trade safeguards (confirmation step, post-trade lockout, item-history tracking).

**Tier 3 — as needed:**

- Behavioral fingerprinting, IP/ASN reputation, community reporting tools.

**Admin tooling built alongside the game, not after.** A bare-bones admin web UI to inspect accounts, view audit logs, ban accounts, and roll back transactions is a v1 requirement.

**Philosophy:** Cheating cannot be made impossible in a browser game. The goal is to make cheating not pay off — server-authoritative state plus detection of obvious bot behavior keeps the field reasonably level.

### 3.7 Persistence Cadence

- **Synchronous, transactional, write-through:** trades, deaths, level-ups, item creation/destruction, account/auth changes.
- **Batched / periodic snapshot:** region state, entity positions, dropped items (every 10–30s while active, on player departure, on graceful shutdown).
- **Not persisted continuously:** player positions during normal play (persisted on logout, region transition, graceful shutdown). On crash, players reappear at last persisted position — the grace-period mechanic handles this.
- **Append-only:** audit log (partitioned by season for cheap drop-at-wipe), chat.

### 3.8 Authentication

- **Email-only login.** Display name is separate and purely cosmetic.
- **bcrypt** for password hashing.
- **Optional TOTP 2FA**, stored encrypted at rest.
- **No SMS-based 2FA** (SIM swap risk).
- **Session tokens hashed before storage** — DB leak does not immediately compromise live sessions.

---

## 4. Database Architecture

### 4.1 Database Choice: PostgreSQL

Chosen because:

- Real ACID transactions for trades and death events.
- JSONB for semi-structured data (procedurally generated item properties, component state, season modifiers).
- Native integer/spatial indexing for tile-grid proximity queries.
- Partitioning for the audit log (drop old partitions at season wipe).
- Mature Go tooling (`pgx`, `sqlc`) which Logan already uses.
- Logical replication and backup story for operational safety.

**Explicitly rejected:**

- SQLite (concurrent write serialization, weaker network access story, awkward wipe operation).
- MongoDB / document DBs (trade subsystem demands relational integrity).
- Graph DBs, time-series DBs (workloads fit Postgres adequately).
- Redis as primary store (not durable; companion only, and deferred).

### 4.2 Operational Approach

- Self-hosted Postgres on Hetzner VPS, single instance.
- Streaming backups (pgBackRest or wal-g) from day one. A DB restore that wipes a week of progress in a hardcore game is community-killing.
- `pgxpool` for connection pooling in the Go process. PgBouncer only if multiple processes ever happen (they shouldn't).
- `pg_stat_statements` enabled from day one for slow-query forensics.
- Migrations via `goose` or `golang-migrate`. Each layer of schema work is a small commit, with forward and back tested on a scratch DB.

### 4.3 Schema-Shaping Principles

**Entity ID type: UUIDv7.** Sortable by creation time, doesn't leak counts, client-generatable for optimistic UI, generated in Go (not in Postgres).

**Tiles are data on a grid, not entities.** A 1000×1000 region as entities would be a million rows of mostly-empty state. Tiles are `(region_id, x, y, tile_type)` data. Things that *happen* to positions (a fire patch, a trap) are entities with positions.

**Items are entities, ownership is a relationship.** Items have a `Location` component: `InInventoryOf(character_id)` | `OnCorpse(corpse_id)` | `OnGroundAt(region_id, x, y)` | `InEscrow(trade_id)` | `InShopStock(shop_id)`. Trades atomically move items between these states inside a Postgres transaction.

**Regions are their own table, not entities.** They're containers/namespaces. Entities have a `region_id` foreign key.

**Component storage: hybrid, lean generic.** Single generic `components(entity_id, component_type, state_jsonb)` table to start. Promote individual component types to dedicated tables only when:

- A specific field needs indexing or query support.
- The component has many rows and needs efficient bulk operations.
- Relational integrity (FKs to other entities) needs to be enforced at the DB level.

**Audit log: action-level granularity** (one row per game action), partitioned by season, append-only. Forensic replay reconstructs state by applying actions from a snapshot.

**What is not persisted (in memory only, lost on restart):**

- WebSocket session state.
- Live cursor / typing / animation / render state.
- FOV calculations, pathfinding caches.

---

## 5. Schema — Layer 1: The Frame

This layer defines accounts, seasons, sessions, and moderation. Outermost layer; everything else hangs off it.

### 5.1 `accounts`

The durable human identity. Survives all seasons.

```sql
CREATE TABLE accounts (
  id              UUID PRIMARY KEY,                       -- UUIDv7
  email           CITEXT NOT NULL UNIQUE,
  email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
  display_name    TEXT NOT NULL UNIQUE,                   -- public name in-game
  password_hash   TEXT NOT NULL,                          -- bcrypt
  totp_secret     TEXT,                                   -- nullable; encrypted at rest
  totp_enabled    BOOLEAN NOT NULL DEFAULT FALSE,
  status          TEXT NOT NULL DEFAULT 'active',         -- active, suspended, banned, deleted
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_login_at   TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ                             -- soft delete
);

CREATE INDEX accounts_status_idx ON accounts (status) WHERE status != 'active';
```

Key decisions:

- `email` is `CITEXT` (case-insensitive). Prevents `Logan@x.com` and `logan@x.com` being treated as different accounts.
- `display_name` separate from email. Email is the login credential; display name is the public identity. Unique because impersonation in a trading game is a real concern.
- `password_hash` with bcrypt. Kept siloed; do not query this column except from the auth path.
- `totp_secret` stored encrypted with an application-level key from env. A DB dump alone should not yield TOTP seeds.
- `status` as text with application-level validation plus a `CHECK` constraint. Avoided Postgres enums due to painful migration semantics.
- Partial index on `status != 'active'` — most rows are active; only the interesting ones get indexed.
- No `username` column. Email-only login. Avoids "I changed my username and now can't log in."

### 5.2 `account_flags`

Sparse, account-level, cross-season state. ECS pattern applied to accounts.

```sql
CREATE TABLE account_flags (
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  flag_type       TEXT NOT NULL,
  flag_value      JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (account_id, flag_type)
);
```

Examples: `cosmetic_unlock`, `achievement`, `staff`, `beta_tester`. Promote a flag type to its own table when it grows large enough (achievements probably will, early).

### 5.3 `seasons`

The wipe boundary, made explicit.

```sql
CREATE TABLE seasons (
  id              INT PRIMARY KEY,                        -- season number: 1, 2, 3...
  name            TEXT,                                   -- optional theme name
  status          TEXT NOT NULL,                          -- upcoming, active, ended
  world_seed      BIGINT NOT NULL,
  modifiers       JSONB NOT NULL DEFAULT '{}'::jsonb,
  starts_at       TIMESTAMPTZ NOT NULL,
  ends_at         TIMESTAMPTZ NOT NULL,
  wiped_at        TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX seasons_one_active ON seasons (status) WHERE status = 'active';
```

Key decisions:

- `id` is `INT`, not UUID. Seasons are human-meaningful and sort naturally.
- Unique partial index enforces "only one active season at a time" at the DB level, not application code.
- `modifiers` as JSONB for per-season rule tweaks without per-season migrations.
- `wiped_at` separate from `ends_at` — there's a window between "season closes" and "world data cleared."

### 5.4 `season_participation`

Per-account, per-season state. Bridges forever-accounts and season-ephemeral characters.

```sql
CREATE TABLE season_participation (
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  season_id       INT NOT NULL REFERENCES seasons(id),
  characters_made INT NOT NULL DEFAULT 0,
  deaths          INT NOT NULL DEFAULT 0,
  deepest_region  INT,                                    -- populated post-mortem
  final_summary   JSONB,                                  -- hall of fame data
  PRIMARY KEY (account_id, season_id)
);
```

### 5.5 `sessions`

Active and recent auth sessions.

```sql
CREATE TABLE sessions (
  id              UUID PRIMARY KEY,                       -- UUIDv7
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  token_hash      TEXT NOT NULL UNIQUE,
  ip_address      INET,
  user_agent      TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at      TIMESTAMPTZ NOT NULL,
  revoked_at      TIMESTAMPTZ,
  revoke_reason   TEXT                                    -- user_logout, admin_revoke, expired, replaced
);

CREATE INDEX sessions_account_active_idx
  ON sessions (account_id) WHERE revoked_at IS NULL;

CREATE INDEX sessions_expires_idx
  ON sessions (expires_at) WHERE revoked_at IS NULL;
```

Key decisions:

- `token_hash`, not the raw token. Same pattern as password storage.
- `ip_address` as `INET` (Postgres native IP type, enables CIDR queries).
- Sessions are not WebSocket connections. A session is the auth token; a WebSocket may reconnect multiple times within one session (grace period mechanic).
- Multiple sessions per account allowed; one live game connection per account enforced in the game server, not the auth layer.
- Periodic cleanup: delete rows where `expires_at < now() - interval '30 days'`.

### 5.6 `moderation_actions`

Append-only audit trail for moderation.

```sql
CREATE TABLE moderation_actions (
  id              UUID PRIMARY KEY,                       -- UUIDv7
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  action_type     TEXT NOT NULL,                          -- ban, suspend, mute, warn, unban
  reason          TEXT NOT NULL,
  details         JSONB NOT NULL DEFAULT '{}'::jsonb,
  applied_by      UUID REFERENCES accounts(id),           -- nullable for system actions
  applied_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at      TIMESTAMPTZ                             -- nullable for permanent actions
);

CREATE INDEX moderation_actions_account_idx
  ON moderation_actions (account_id, applied_at DESC);

CREATE INDEX moderation_actions_active_idx
  ON moderation_actions (account_id, expires_at) WHERE expires_at IS NOT NULL;
```

Key decisions:

- Append-only. Never update or delete. Unbanning is a new row with `action_type = 'unban'`.
- `accounts.status` is a denormalized cache. This table is the source of truth.
- `applied_by` nullable for automated actions (rate-limit auto-suspends, etc.).

### 5.7 Explicitly Deferred from Layer 1

- Password reset flow (`password_reset_tokens` table) — add when wiring up signup.
- Email verification flow (`email_verification_tokens` table) — same.
- OAuth / SSO (`account_oauth_links` table) — add later if ever.
- Per-account UI preferences (keybinds, settings) — start in `account_flags`, promote if they grow.
- Friends list, blocklist — social graph, separate design conversation.

### 5.8 Layer 1 Commit Sequence

1. Migrations infrastructure (`goose` set up, empty up/down) — verify the migration tool works on the dev DB.
2. Extensions enabled (`citext`).
3. `accounts` + `account_flags` tables.
4. `seasons` + `season_participation` tables, plus seed migration creating season 1.
5. `sessions` table.
6. `moderation_actions` table.
7. Go types and `sqlc` queries for basic operations (create account, look up by email, create session, validate session, list active sessions).

---

## 6. Schema — Layer 2: Entities and Components

This layer defines the universal "thing in the world" abstraction. Every game object — character, NPC, item, corpse, projectile, world object — is an entity. Behavior and state are layered on via components plus per-type tables (Layer 3+).

### 6.1 ECS Philosophy (Clarified)

Three approaches were considered:

- **Pure ECS** — no entity types, identity is emergent from component set.
- **Pure ECS in storage** — generic `(entity_id, component_type, state_jsonb)` for everything, marker components for "is a character."
- **Hybrid (chosen)** — entities are compositional in *thinking* (a character is "an entity with Position, Health, Inventory, Actor components plus a row in `characters`"), but storage is hybrid: per-type normalized tables for core attributes (Layer 3), generic components table for sparse extensible state.

`entity_type` is a **discriminator** for fast lookup, not the *definition* of what the entity is. The definition remains compositional. This matches §2.4's "ECS-style discipline applied to conventional Go architecture."

### 6.2 `entities`

The universal handle. Every game object has one row here.

```sql
CREATE TABLE entities (
  id                UUID PRIMARY KEY,           -- UUIDv7
  season_id         INT NOT NULL REFERENCES seasons(id),
  entity_type       TEXT NOT NULL,
  created_at_tick   BIGINT NOT NULL,
  destroyed_at_tick BIGINT,                     -- nullable; soft delete
  CHECK (entity_type IN (
    'character', 'npc', 'item', 'corpse', 'projectile', 'world_object'
  ))
);

CREATE INDEX entities_season_type_idx
  ON entities (season_id, entity_type);

CREATE INDEX entities_destroyed_idx
  ON entities (destroyed_at_tick)
  WHERE destroyed_at_tick IS NOT NULL;
```

Key decisions:

- **UUIDv7 IDs**, generated in Go. Sortable by creation, client-generatable for optimistic UI.
- **`entity_type` is TEXT with a CHECK constraint**, not a lookup table. Closed set defined by code, not content. Mirrors the `accounts.status` pattern.
- **`entity_type` is immutable.** Character-to-corpse is "destroy character entity, create corpse entity" — two IDs, clean lifecycle. No DB trigger enforces this; application code is the source of truth.
- **Six types in v1**: `character`, `npc`, `item`, `corpse`, `projectile`, `world_object`. Coarse on purpose — finer distinctions (fire patch vs. trap vs. door) live in components and templates.
- **`projectile` is a first-class type.** Arrows in flight, magic missiles, thrown bombs, explosives with fuses — all entities with Position + Actor + a `Trajectory` component. The scheduler ticks them the same way it ticks players and NPCs. This unlocks interceptable, dodgeable, lootable-on-miss projectiles without special-case loops.
- **Soft delete via `destroyed_at_tick`.** A destroyed entity is inert but still queryable for audit-log replay. Hard-delete sweep runs periodically (retention window ~24–48 game-hours); full truncation at season wipe.
- **No `destroyed_reason` column.** Reasons live in the audit log (Layer 6).
- **`season_id` FK** ensures entities are scoped to a season for cheap wipe.

### 6.3 `entity_positions`

Sparse table for entities that have a position in the world. Most entities have one; some abstract or container-held entities don't.

```sql
CREATE TABLE entity_positions (
  entity_id        UUID PRIMARY KEY REFERENCES entities(id) ON DELETE CASCADE,
  region_id        INT NOT NULL,                -- FK added in Layer 4
  x                INT NOT NULL,
  y                INT NOT NULL,
  updated_at_tick  BIGINT NOT NULL
);

CREATE INDEX entity_positions_region_xy_idx
  ON entity_positions (region_id, x, y);

CREATE INDEX entity_positions_region_idx
  ON entity_positions (region_id);
```

Key decisions:

- **Position is its own table, not columns on `entities` or a component in JSONB.** Hot-path spatial queries deserve a real indexed table. Not nullable-columns-on-entities because most entities have a position but the abstract ones don't, and nullable columns lie about that. Not a JSONB component because position has no schema variation and is queried spatially, not structurally.
- **Position row deleted when entity is destroyed.** Position is liveness state. A destroyed entity isn't in the world. Cascade delete on `ON DELETE CASCADE`; for soft-deletes, application code deletes the position row when setting `destroyed_at_tick`.
- **`region_id` FK deferred to Layer 4.** Layer 2's migration creates the column without the FK; Layer 4 adds `ALTER TABLE ... ADD FOREIGN KEY`. Keeps layer boundaries honest.
- **Composite index `(region_id, x, y)`** for spatial queries — "what's at (x,y) in region R" and "what's in this region."
- **`updated_at_tick`** enables broadcast debouncing and grace-period reconciliation.

### 6.4 `components`

Generic sparse storage for optional, extensible per-entity state.

```sql
CREATE TABLE components (
  entity_id        UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  component_type   TEXT NOT NULL,
  state            JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at_tick  BIGINT NOT NULL,
  updated_at_tick  BIGINT NOT NULL,
  PRIMARY KEY (entity_id, component_type)
);

CREATE INDEX components_type_entity_idx
  ON components (component_type, entity_id);
```

Key decisions:

- **Composite PK `(entity_id, component_type)`** enforces one component of each type per entity. If multiple are needed, that's a signal to promote that component to a dedicated table.
- **`component_type` is free-form text, validated in Go.** Component types are a closed set defined by code (typed structs serialized to JSONB). Content (item/monster templates) composes existing components, never defines new ones. A CHECK constraint listing known component types is a future possibility once the set stabilizes.
- **No JSONB GIN indexes yet.** Add when a system needs to query into component state — and treat that as a signal to promote the component to its own table.
- **Components are current state, not history.** Mutable in place. History lives in the audit log.
- **Cascade delete on entity removal**, but soft-deleted entities retain their components until the periodic hard-delete sweep. Application queries filter on `entities.destroyed_at_tick IS NULL` explicitly when iterating components.
- **`created_at_tick` and `updated_at_tick`** for forensic queries and broadcast debouncing.

### 6.5 Explicitly Deferred from Layer 2

- **`action_costs` table** — gameplay data for the energy system. Deferred to Layer 3 when there are abilities to assign costs to.
- **Per-type tables** (`characters`, `items`, `corpses`, etc.) — Layer 3.
- **`Actor` component** (energy, regen, next-ready-tick) — Layer 3.
- **`Location` component variants** (InInventoryOf, OnCorpse, OnGroundAt, InEscrow, InShopStock) — Layer 3.
- **`Trajectory` / `InFlight` component shape, impact resolution, decay-to-item-on-miss** — Layer 3.
- **Entity templates / prefabs** — Layer 3 or its own concern.
- **Promotion of specific component types to dedicated tables** — only when a real query, indexing, or integrity need arises.
- **Spatial indexing beyond btree on `(region_id, x, y)`** — GiST/BRIN deferred until measured need.
- **Position update batching** — currently every move is one UPDATE. If we measure hot-spot contention, batch at the region-goroutine level. Defer.
- **Server-side `current_tick` persistence cadence** — small concern, decide when wiring up the scheduler.

### 6.6 Layer 2 Commit Sequence

1. **Migration: create `entities` table** with CHECK constraint and indexes. No FKs to it from other tables yet.
2. **Migration: create `entity_positions` table** with indexes. `region_id` as plain `INT NOT NULL`, no FK (Layer 4 adds it).
3. **Migration: create `components` table** with indexes.
4. **Go types**: `Entity`, `Position`, generic `Component` interface, and the first concrete component struct as a smoke test (something cheap like a `Hidden` marker — real ones like `OnFire` arrive in Layer 3 when systems exist to use them).
5. **`sqlc` queries for basic operations**:
   - Create entity (with optional initial position and components, in a transaction)
   - Look up entity by ID (with components, with position)
   - Soft-delete entity (set `destroyed_at_tick`, delete position row)
   - Query entities by `(season_id, entity_type)` filter
   - Query entities by `(region_id, x, y)` and `(region_id)` via position
   - Get/set/delete a component on an entity
   - Iterate entities with a given component type (joined with `entities.destroyed_at_tick IS NULL`)
6. **Sweep job stub** for hard-deleting entities where `destroyed_at_tick < (current_tick - retention_window)`. Wired but disabled by config until Layer 6 (audit log) can confirm no references exist.

---

## 7. Schema — Layer 3: Characters and Items

This layer puts flesh on the entity/component frame. It is the largest layer by decision count and is being committed in pieces. Each piece is its own conversation and its own commit boundary.

**Layer 3 piece status:**

- **Characters** — designed; §7.2.
- **Actors** — designed; §7.3.
- **Items + Location component** — pending.
- **Corpses + death-and-loot flow** — pending.
- **`action_costs` table** — pending.
- **Projectile mechanics** (`Trajectory` component, impact resolution) — pending.
- **Entity templates / prefabs** — pending.

### 7.1 Refinements to the ECS-Hybrid Model

Layer 3 surfaces two refinements to §2.4 / §6.1 worth stating explicitly:

**Promotion-on-day-one is acceptable when the §4.3 criteria are met immediately.** `actors` is the first example: every character, NPC, and projectile has one; the scheduler queries `next_ready_tick` every tick across thousands of rows. Storing this in the generic `components` table as JSONB and extracting on every scheduler pass would be wasteful. A dedicated table with a btree index is the right tool from day one. This isn't a retreat from the hybrid model — it's the model working as intended. The decision rule is the §4.3 criteria, not "always start in `components`."

**Defining state is core, modifying effects are components.** A character's HP is *defining*: every character has it, it's queried hot, and it changes shape only with major design work. HP belongs as columns on `characters`. A character's `OnFire` status is *modifying*: sparse, temporary, optional, evolving. It belongs in `components`. This rule resolves the "is HP a Health component?" question and provides a precedent for similar decisions in later pieces of Layer 3.

### 7.2 `characters`

The per-season character record for a player-controlled entity.

```sql
CREATE TABLE characters (
  entity_id       UUID PRIMARY KEY REFERENCES entities(id) ON DELETE CASCADE,
  account_id      UUID NOT NULL REFERENCES accounts(id),
  season_id       INT NOT NULL REFERENCES seasons(id),
  character_name  CITEXT NOT NULL,
  hp              INT NOT NULL,
  hp_max          INT NOT NULL,
  created_at_tick BIGINT NOT NULL,
  died_at_tick    BIGINT
);

CREATE UNIQUE INDEX characters_season_name_idx
  ON characters (season_id, character_name);

CREATE INDEX characters_account_active_idx
  ON characters (account_id, season_id)
  WHERE died_at_tick IS NULL;

CREATE INDEX characters_hp_zero_idx
  ON characters (entity_id)
  WHERE hp <= 0 AND died_at_tick IS NULL;
```

Key decisions:

- **`entity_id` is both PK and FK to `entities`.** A character is a subtype of entity. Cascade delete on entity destruction handles cleanup of the character row, position, components, and (later) actor row in one fall.
- **`account_id` and `season_id` denormalized** from the relationship through `entities.season_id`. Cheap, and makes "find this account's character" and "all characters this season" fast without joining `entities` first.
- **`character_name` is distinct from `accounts.display_name`.** The account display name identifies the *player* persistently across seasons; the character name identifies the in-world *character* who will eventually die. Both are visible in-game. Death feed reads "Grimsnarl the Brave (Logan) was killed by a hill troll." One account display name maps to many character names over time.
- **`character_name` uniqueness is per-season.** Within a season, no two living characters share a name — prevents in-game confusion. Across seasons, names recycle freely. No persistent claim table. Wipe clears character names along with characters.
- **`character_name` is `CITEXT`.** Case-insensitive comparison. "Grimsnarl" and "grimsnarl" are the same name.
- **`hp` and `hp_max` are columns, not a `Health` component.** Reasoning: hot-path read-modify-write on every damage event; indexable death-detection query; HP is *defining* state for characters. Transient effects that modify HP (`OnFire`, `Poisoned`, `Regenerating`) live as components — they affect HP but they aren't HP.
- **`died_at_tick` is denormalized from `entities.destroyed_at_tick`.** Semantically clearer in the characters context and queryable without joining. Set in the same transaction that sets `entities.destroyed_at_tick`.
- **Partial index on living characters per account** (`WHERE died_at_tick IS NULL`). The hot query is "find this player's currently living character for this season" — runs on every player connection. Partial index keeps it cheap as the season's dead-character count grows.
- **Partial index on `WHERE hp <= 0 AND died_at_tick IS NULL`** — the death-detection sweep. Usually empty (characters with non-positive HP get processed and marked dead within a tick), near-zero cost when empty.
- **No `is_active` flag.** Hardcore permadeath means an account has at most one living character per season; `WHERE account_id = X AND season_id = current AND died_at_tick IS NULL` is the predicate. No denormalized flag needed.

### 7.3 `actors`

The energy-and-scheduling state for any entity that takes actions over time.

```sql
CREATE TABLE actors (
  entity_id        UUID PRIMARY KEY REFERENCES entities(id) ON DELETE CASCADE,
  energy           INT NOT NULL,
  energy_regen     INT NOT NULL,
  energy_cap       INT NOT NULL,
  next_ready_tick  BIGINT,             -- NULL means "not scheduled"
  updated_at_tick  BIGINT NOT NULL
);

CREATE INDEX actors_ready_idx
  ON actors (next_ready_tick)
  WHERE next_ready_tick IS NOT NULL;
```

Key decisions:

- **Promoted to a dedicated table on day one.** Per §7.1: the scheduler queries `next_ready_tick` every tick across potentially thousands of rows. JSONB extraction is the wrong tool. A btree index on a plain BIGINT is the right tool. `actors` is the precedent the §7.1 promotion rule points back to.
- **`entity_id` PK-and-FK.** Same pattern as `characters`. Cascade delete is clean.
- **Sparse across entity types.** Characters, NPCs, and projectiles have rows; corpses, ground-items, and static world objects don't. This sparseness is the structural reason the dedicated table works — we're not storing zero rows for entities that don't act.
- **`energy`, `energy_regen`, `energy_cap` as plain INT columns.** Hot-path read-modify-write. JSONB would be wasteful.
- **`next_ready_tick` is a denormalized cache.** Strictly redundant — derivable from `energy + energy_regen * (current_tick - updated_at_tick)` — but the scheduler runs every tick and needs to answer "who's ready?" cheaply. Cache updated on the rare path (action taken, regen modified, status effect applied/removed); read on the hot path (every scheduler tick).
- **`next_ready_tick` is nullable; NULL means "not scheduled."** An actor with full energy and no pending intent gets `next_ready_tick = NULL`. The partial index `WHERE next_ready_tick IS NOT NULL` excludes idle actors from the scheduler's hot query entirely. The scheduler reads `SELECT entity_id FROM actors WHERE next_ready_tick <= $current_tick ORDER BY next_ready_tick` — O(log n) on the ready set, not on all actors.
- **Re-enabling scheduling is a side-channel.** Player input arrival, NPC perception event, projectile spawn — any code path that creates intent must set `next_ready_tick`. Contained through a single helper (`MarkActorReady(entityID, tick)`); all intent-creation paths route through it. Code review catches a path that doesn't.
- **`updated_at_tick`** supports the lazy energy-recompute fallback (true current energy is `energy + energy_regen * (current_tick - updated_at_tick)`) and forensic queries.
- **No `current_intent` column.** Intent is volatile and changes on every input. The scheduler only asks "are you ready?"; a separate in-memory queue keyed by entity_id holds pending actions. If crash-recoverable queued actions ever matter, that's a future `actor_intents` table — not this one.
- **No `region_id` column yet.** Region-scoped scheduler queries currently join through `entity_positions`. If sharding across regions ever happens (§2.2 defers this), we'd denormalize `region_id` onto `actors` for cheaper per-region scheduling.

### 7.4 Layer 3 Piece 1 Commit Sequence (characters + actors)

1. **Migration: create `characters` table** with indexes.
2. **Migration: create `actors` table** with the partial scheduling index.
3. **Go types**: `Character`, `Actor` structs with `sqlc`-friendly tagging.
4. **`sqlc` queries for characters**:
   - Create character (in a transaction that also inserts the `entities` row, an `entity_positions` row, and an `actors` row).
   - Look up living character for an account in the current season.
   - Apply damage (UPDATE `characters` SET hp = hp - $1).
   - Mark character dead (set `died_at_tick` and `entities.destroyed_at_tick` in the same transaction; delete `entity_positions` row).
   - Query dead-but-not-yet-processed characters (`hp <= 0 AND died_at_tick IS NULL`) for the death sweep.
5. **`sqlc` queries for actors**:
   - Insert actor row at character/NPC/projectile creation.
   - `MarkActorReady(entity_id, tick)` — sets `next_ready_tick`.
   - Scheduler query — entities ready to act, ordered by `next_ready_tick`.
   - Deduct energy and recompute `next_ready_tick` (or set NULL if energy is full and no pending intent).
6. **Helper function in Go**: `MarkActorReady` is the single chokepoint for re-enabling scheduling. All intent-creation paths route through it.

### 7.5 Explicitly Deferred from Layer 3 Piece 1

- **Items + Location component** — next piece.
- **Corpses + death-and-loot flow** — depends on items being concrete.
- **`action_costs` table** — small piece, can come anytime after items.
- **Projectile mechanics** (`Trajectory` component, impact resolution, lootable-on-miss) — own piece.
- **Entity templates / prefabs** — own piece, debate over DB-table vs. file-based.
- **Equipment slots** — likely a JSONB component initially; deferred until items exist.
- **Status effects** (`OnFire`, `Poisoned`, `Stunned`, `Regenerating`) — components in the generic table; defined when systems need them.
- **XP / level / skills** — combat-system concern, possibly post-MVP.
- **Hunger, thirst, fatigue** — survival-system concern, post-MVP.
- **Faction / alignment / reputation** — social-system concern, post-MVP.

---

## 8. Remaining Schema Layers

The schema work is structured as six layers. Each is a natural conversation and a real commit boundary.

1. **The frame** — accounts, seasons, sessions, moderation. *(designed; §5)*
2. **Entities and components** — the shared "thing in the world" abstraction. *(designed; §6)*
3. **Characters and items** — most-trafficked entity subtypes. *(in progress; §7. Characters and actors designed. Items + Location, corpses, action_costs, projectile mechanics, templates pending.)*
4. **World structure** — regions, tiles, world generation, spatial indexing. Adds the `entity_positions.region_id` FK.
5. **Trades** — the hardest atomic operation, its own table family.
6. **Audit log and chat** — append-only, partitioned, forensic-ready.

---

## 9. Open Questions and Future Decisions

Captured here so they don't get lost. Not blocking near-term work.

### Design questions

- Region size (tile dimensions per region)?
- Exact idle timeout duration in unsafe zones?
- Exact grace period duration on disconnect?
- World size — how many regions in a season?
- "Danger / loot tier" mechanic — discrete tiers, continuous gradient, or something else?
- Mobile / touch support — in scope or desktop-only?
- Visual style — strict ASCII, tilemap, sprites?

### Technical questions

- Region transition handshake — how does a player physically move from region A to region B?
- Message protocol for the WebSocket layer (action types, event types, state delta shape).
- FOV algorithm choice (shadowcasting, raycasting, etc.).
- Pathfinding for NPCs (A*, JPS, flow fields).
- Client rendering approach (DOM grid vs. canvas, decided in frontend layer discussion).
- Tick rate value (100ms? 250ms? other?) — server config, tunable in deployment.
- Default action costs (move, attack, throw, cast) — seeded by migration in Layer 3.

### Operational questions

- Monitoring stack (Logan already runs Grafana/Prometheus/Loki — adapt existing setup).
- Deployment pipeline (CI/CD via GitHub Actions to Hetzner — existing pattern).
- How to test the season wipe operation safely.
- Backup verification cadence.
- `current_tick` persistence cadence (every N ticks? on graceful shutdown only?).

---

## 10. Changelog

- **2026-05-15 (initial)** — Initial document. Layer 1 (the frame) designed. Conceptual model, multiplayer design decisions, database choice, and ECS-hybrid runtime model locked in. Working title set to "Walking Drum."
- **2026-05-15 (Layer 2)** — Time system established (§2.5): energy-based, server-driven, tick rate as network-layer concern, action costs as gameplay data, game-tick as source of truth for in-world time. Layer 2 (entities and components) designed (§6): `entities` table with six-type taxonomy including `projectile`, `entity_positions` as a sparse dedicated table, `components` as generic JSONB. ECS philosophy clarified as hybrid — compositional thinking, normalized storage where it pays. Layer numbering updated; remaining layers shifted accordingly.
- **2026-05-16 (Layer 3 — piece 1: characters and actors)** — First piece of Layer 3 designed (§7). ECS-hybrid model refined (§7.1) with two rules: promotion-on-day-one is acceptable when §4.3 criteria are met immediately; defining state is core, modifying effects are components. `characters` table defined (§7.2) with HP as columns (not a Health component), `character_name` distinct from account display name, character names unique per-season, and partial indexes for the active-character lookup and death-detection hot paths. `actors` table defined (§7.3) as a day-one promotion with `next_ready_tick` nullable as the "not scheduled" off-switch, partial index keeping the scheduler query cheap. Remaining Layer 3 pieces (items + Location, corpses, action_costs, projectile mechanics, templates) deferred to follow-up conversations.