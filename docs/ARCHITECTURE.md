# Architecture

## Design

The application is a Go modular monolith with SQLite persistence, a local
SQLite mock fixture, and a JSON HTTP API. The MVP processes batches
sequentially.

## Project layout

```text
cmd/catalogd/
internal/catalog/domain/
internal/catalog/service/
internal/catalog/repository/sqlite/
internal/catalog/httpapi/
examples/mock-data.db
docs/
```

## Package responsibilities

- `domain`: catalog models, request models, and shared domain errors. It has no
  HTTP or database dependencies.
- `service`: update orchestration, validation, version policy, issue handling,
  seed coordination, and use-case tests.
- `repository/sqlite`: SQLite schema, versioned migrations, row mapping, and
  transactional persistence.
- `httpapi`: route registration, request decoding, response mapping, and HTTP
  tests. It depends on the service and domain packages, not SQL.
- `cmd/catalogd`: application assembly through the Store and Handler factories,
  flags, HTTP server lifecycle, and graceful shutdown.

## Dependency direction

```text
cmd/catalogd -> httpapi -> service -> domain
                           service -> repository/sqlite -> domain
```

The domain package must remain independent. HTTP handlers do not execute SQL
or implement catalog version rules.

## Current persistence boundary

The service owns the in-memory state snapshot and serializes state changes.
The SQLite repository loads and commits that snapshot transactionally. Products
are stored in a queryable table; update reports currently store nested batches,
changes, and issues as JSON. Future migrations can normalize those records
without changing HTTP routing.

The seed command reads the repository's `examples/mock-data.db` fixture in
read-only mode and imports its products through the service. Demo event data is
kept in the fixture but is not yet a runtime supplier adapter.

## Transactions and recovery

State changes are written in one SQLite transaction. The in-memory snapshot is
updated only after a successful commit. On restart, previously in-progress
updates are marked interrupted and can be resumed. A repeated batch ID with the
same payload remains idempotent.

## Concurrency and limitations

The service serializes changes within one process. SQLite has a busy timeout,
but cross-process update ownership is not yet implemented. Do not run multiple
service processes against the same database. Event receipts, attempts, and
product history are not yet normalized into independent tables.
