# Data Model

## Conventions

SQLite with foreign keys enabled, versioned migrations, UTC timestamps,
integer `price_minor` (never float), persistent tombstones, no
hard-delete of history/receipts.

## Tables

### products

-   `product_id TEXT PRIMARY KEY`
-   `name TEXT NOT NULL`
-   `description TEXT NOT NULL DEFAULT ''`
-   `price_minor INTEGER CHECK(price_minor >= 0)`
-   `currency TEXT`
-   `stock_quantity INTEGER CHECK(stock_quantity >= 0)`
-   `status TEXT CHECK(status IN ('ACTIVE','WITHDRAWN'))`
-   `source_version INTEGER NOT NULL CHECK(source_version > 0)`
-   `last_event_id TEXT`
-   `created_at TEXT NOT NULL`, `updated_at TEXT NOT NULL`

For a withdrawn product retain the row and source_version as a
tombstone.

### updates

-   `id TEXT PRIMARY KEY` (UUID)
-   `status TEXT NOT NULL`
-   `source_name TEXT NOT NULL`
-   `input_path TEXT NOT NULL` (internal managed path, not arbitrary
    client path)
-   `input_sha256 TEXT NOT NULL`
-   `created_at TEXT NOT NULL`, `started_at TEXT NULL`,
    `finished_at TEXT NULL`
-   `last_error_code TEXT NULL`, `last_error_message TEXT NULL`

Enforce one active update using a transactional guard/lock strategy; a
non-atomic check is insufficient.

### batches

-   `id TEXT PRIMARY KEY`
-   `update_id TEXT NOT NULL REFERENCES updates(id)`
-   `batch_index INTEGER NOT NULL`
-   `status TEXT NOT NULL`
-   `input_path TEXT NOT NULL`, `input_sha256 TEXT NOT NULL`
-   `attempt_count INTEGER NOT NULL DEFAULT 0`
-   `started_at TEXT NULL`, `finished_at TEXT NULL`
-   `last_error_code TEXT NULL`, `last_error_message TEXT NULL`
-   `UNIQUE(update_id, batch_index)`

### event_receipts

Stores the canonical result for a stable event ID: -
`event_id TEXT PRIMARY KEY` - `update_id`, `batch_id` foreign keys -
`product_id TEXT NULL`, `source_version INTEGER NULL` -
`payload_sha256 TEXT NOT NULL`, `action TEXT NULL`,
`outcome TEXT NOT NULL` - `processed_at TEXT NOT NULL`

If the same event ID appears with a different payload, do not overwrite
the original receipt. Store the conflicting occurrence separately and
create an issue.

### event_attempts

-   `id TEXT PRIMARY KEY`
-   `update_id`, `batch_id` foreign keys
-   `event_index INTEGER NOT NULL`
-   `event_id TEXT NULL`, `product_id TEXT NULL`
-   `payload_sha256 TEXT NOT NULL`, `outcome TEXT NOT NULL`
-   `issue_id TEXT NULL`, `created_at TEXT NOT NULL`
-   `UNIQUE(batch_id, event_index)` so replay of an immutable batch does
    not duplicate the same event occurrence.

### product_history

-   `id TEXT PRIMARY KEY`
-   `product_id TEXT NOT NULL`
-   `update_id`, `batch_id` foreign keys
-   `event_id TEXT NOT NULL`, `source_version INTEGER NOT NULL`
-   `change_type TEXT NOT NULL`
-   `before_json TEXT NULL`, `after_json TEXT NOT NULL`
-   `created_at TEXT NOT NULL`
-   Unique key preventing duplicate history for one applied event.

### issues

-   `id TEXT PRIMARY KEY`
-   `update_id` FK, `batch_id TEXT NULL`
-   `event_id TEXT NULL`, `product_id TEXT NULL`
-   `code TEXT NOT NULL`, `severity TEXT NOT NULL`,
    `message TEXT NOT NULL`
-   `dedupe_key TEXT NOT NULL UNIQUE`
-   `status TEXT NOT NULL CHECK(status IN ('OPEN','RESOLVED'))`
-   `created_at TEXT NOT NULL`, `resolved_at TEXT NULL`

Keep issues after resolution; resolving must not erase audit history.

### schema_migrations

`version INTEGER PRIMARY KEY`, `applied_at TEXT NOT NULL`.

## Indexes

Index `products(status, updated_at)`, `updates(status, created_at)`,
`batches(update_id, status, batch_index)`,
`product_history(product_id, created_at)`,
`issues(update_id, status, created_at)`.

## Atomicity

For a product mutation, write product + event outcome/receipt + history
in one transaction. For reject/conflict, persist event attempt/outcome +
issue together. Mark a batch completed only after every event has a
committed outcome. Progress counters must be derived from persistent
rows or updated atomically with deduplication.
