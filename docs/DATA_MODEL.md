# SQLite Data Model

## Database setup

The application creates the SQLite database on startup and applies
versioned, idempotent migrations. The default database path is
`./data/catalog.db`; the parent directory is created automatically. The
database file is restricted to owner read/write permissions.

## Current tables

### `products`

The catalog's queryable source of truth:

- `product_id TEXT PRIMARY KEY`
- `name TEXT NOT NULL`
- `description TEXT NOT NULL DEFAULT ''`
- `price_cents INTEGER NOT NULL CHECK(price_cents >= 0)`
- `revision INTEGER NOT NULL CHECK(revision > 0)`
- `withdrawn INTEGER NOT NULL DEFAULT 0`

Withdrawals retain the product row as a tombstone. An index supports status
and revision filtering.

### `updates`

- `id TEXT PRIMARY KEY`
- `status TEXT NOT NULL`
- `body_json TEXT NOT NULL`

The JSON body currently stores the update report, its batches, changes, and
issues. A status index supports filtering. These nested records are persisted
atomically with the product table, but are not yet independently queryable.

### `schema_migrations`

- `version INTEGER PRIMARY KEY`
- `applied_at TEXT NOT NULL`

Migration version 1 creates the catalog and update tables and indexes. Reopening
the database does not reapply the schema migration.

## Write atomicity and recovery

Each state change runs in a SQLite transaction. Product rows and update report
rows are replaced together, then the in-memory state is updated only after the
database commit succeeds. On startup, the application reloads committed rows
and marks previously `in_progress` updates as `interrupted` so they can be
resumed.

## Seed data

`examples/mock-data.db` contains 100 deterministic fictional products and the
demo event records. Run `go run ./cmd/catalogd -data ./data/catalog.db -seed
./examples/mock-data.db` to import the products. The fixture is opened in
read-only mode. Matching existing products are skipped; conflicting IDs fail
without overwriting existing catalog data.

## Planned normalization

The specification describes separate batch, event receipt, event attempt,
product history, and issue tables. Those tables are not part of the current
incremental implementation. Add them with new migrations before relying on
durable per-event querying or multi-instance worker coordination.
