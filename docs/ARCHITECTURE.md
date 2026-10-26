# Architecture

## Design

Modular monolith in Go; SQLite persistent storage; JSON fixture input;
HTTP JSON API. Process batches sequentially in MVP. Avoid goroutines for
batch processing until correctness and tests are established.

## Suggested layout

``` text
cmd/catalogd/main.go
internal/config/
internal/domain/
internal/input/
internal/service/
internal/repository/
internal/repository/sqlite/
internal/httpapi/
internal/observability/
migrations/
examples/
docs/
.github/
```

## Responsibilities

-   `domain`: Product, SupplierEvent, Update, Batch, Issue, EventOutcome
    and domain rules. Must not import HTTP/SQLite.
-   `input`: fixture allowlist, parsing, immutable snapshot and
    checksum.
-   `service`: update orchestration, validation, version policy,
    batch/event processing.
-   `repository`: interfaces only.
-   `repository/sqlite`: SQL, migrations and transactions.
-   `httpapi`: request parsing, response mapping, status codes; no
    SQL/business rules.
-   `config`: env config and startup validation.

## Flow

1.  Validate fixture ID against allowlist.
2.  Parse input and create immutable snapshot + SHA-256.
3.  In transaction, create update and batches; enforce one active
    update.
4.  Claim pending/retryable batch using conditional state transition.
5.  For each event, validate and process inside its own transaction.
6.  Commit product mutation + receipt/attempt + history OR outcome +
    issue atomically.
7.  Mark batch final only after every event has a persistent outcome.
8.  Recompute update status from persisted batch states.
9.  On restart, reconcile stale RUNNING state and resume only unfinished
    work.

## Transactions and recovery

Each event is a transaction so one invalid event does not rollback the
whole batch. If commit succeeded but acknowledgement was lost, replay
must detect the persistent receipt/attempt and avoid side effects twice.
Never clear catalog before reading/validating new input. Never mark
batch complete before event outcomes are committed.

## Error policy

-   Validation/conflict: persistent issue, no automatic retry.
-   Transient DB/IO: retryable failure or interrupted state.
-   Permanent snapshot/checksum failure: failed state with safe message.
-   Never swallow errors or expose stack traces through API.

## Concurrency

MVP is sequential but still enforce unique product IDs, atomic
conditional state transitions and one active update in transaction. If
parallel workers are later added, require leases/owner tokens and
ownership checks before writes; add race/concurrency tests before
enabling.

## Shutdown and logging

Use context cancellation and graceful HTTP shutdown. Stop accepting new
updates, allow in-flight transaction to finish where possible, and leave
unfinished work resumable. Structured logs should include
update_id/batch_id/event_id when available. Logs are not a substitute
for persistent issues.
