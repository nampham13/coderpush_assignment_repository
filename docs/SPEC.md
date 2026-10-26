# Product Specification

## Goal

Staff can start a catalog update, review its progress, changes, and issues,
and resume an interrupted update after an application restart. Demo data is
entirely fictional.

## Scope

**In scope:** Go HTTP API, SQLite, JSON fixtures, batches, validation,
create/update/withdraw/reactivate operations, version ordering, persistent
issues and history, retry/resume, tests, and a demo.

**Out of scope for the MVP:** real supplier integrations, authentication, a
complete frontend, distributed workers, parallel batch processing, and
inferring a withdrawal from a product missing from a batch.

## Event schema

```json
{
  "event_id": "evt-001",
  "product_id": "SKU-100",
  "source_version": 3,
  "action": "UPSERT",
  "name": "Widget",
  "description": "Fictional item",
  "price_minor": 1299,
  "currency": "USD",
  "stock_quantity": 15
}
```

- `event_id`: stable event identifier.
- `product_id`: trim surrounding whitespace; must not be empty; case-sensitive.
- `source_version`: positive integer that increases monotonically per product.
- `action`: `UPSERT`, `WITHDRAW`, or `REACTIVATE`.
- `UPSERT` and `REACTIVATE` require a non-empty name, non-negative
  `price_minor`, a three-letter uppercase currency, and non-negative
  `stock_quantity`. `description` defaults to an empty string.
- `WITHDRAW` requires only the event ID, product ID, version, and action.
- Unknown actions, missing fields, malformed types, and invalid values create
  persistent issues; they must not terminate the entire update.

## Required invariants

1. Each product ID identifies at most one product.
2. An older version never overwrites a newer version.
3. An older event cannot restore a withdrawn product.
4. Only a valid `REACTIVATE` event with a higher version can reactivate a
   product.
5. Replaying input does not duplicate mutations, receipts, history, or issues.
6. Product mutation, event outcome, and history are committed in one
   transaction.
7. An update is not complete while batches remain unprocessed or policy-defined
   errors remain unresolved.
8. An invalid event does not remove existing catalog data.
9. A product missing from a batch is not considered withdrawn.
10. The input snapshot remains immutable across retry and resume operations.
11. A different payload or action at the same version is a
    `VERSION_CONFLICT`; current state is retained.
12. A different payload using the same event ID is an `EVENT_ID_CONFLICT`;
    the original event is retained.

## Version policy

- A previously unseen product can be created by `UPSERT`.
- A valid event with a greater version applies its action.
- A lower version is `STALE_IGNORED` and does not mutate the product.
- The same version with the same payload and action is `DUPLICATE_NOOP`.
- The same version with a different payload or action raises a conflict issue
  and preserves current state.
- `WITHDRAW` retains a tombstone and its source version.
- `REACTIVATE` is rejected when the product has never existed in the MVP.
- Compare source versions per product; do not use local timestamps for ordering.

## States

Batch states: `PENDING`, `RUNNING`, `COMPLETED`, `COMPLETED_WITH_ISSUES`,
`RETRYABLE_FAILED`, and `FAILED`.

Update states: `PENDING`, `RUNNING`, `COMPLETED`,
`COMPLETED_WITH_ISSUES`, `INTERRUPTED`, and `FAILED`. Add `CANCELLED` only if
a cancel API is introduced.

- `COMPLETED`: every batch is finished with no issue or error.
- `COMPLETED_WITH_ISSUES`: every batch reached a terminal state, but one or
  more events were rejected or conflicted.
- `INTERRUPTED`: resumable work remains.
- `FAILED`: a permanent error prevents continuation.
- Enforce at most one active update transactionally, not only in the UI.
- Retry transient failures; do not retry validation errors or conflicts without
  a bound.
- A batch completes only after every event outcome is committed.

## Progress and issues

Event outcomes are `CREATED`, `UPDATED`, `WITHDRAWN`, `REACTIVATED`,
`UNCHANGED`, `DUPLICATE_NOOP`, `STALE_IGNORED`, `REJECTED`,
`VERSION_CONFLICT`, and `EVENT_ID_CONFLICT`.

Persist each issue's ID, update/batch/event/product associations, code,
severity, message, `OPEN`/`RESOLVED` status, timestamps, and deduplication key.
Derive progress from persistent state so retries do not inflate counts.

## Acceptance criteria

- Newer versions win; stale versions are ignored.
- Duplicate events do not duplicate products or history.
- A stale `UPSERT` cannot undo a withdrawal.
- Invalid records create issues while valid records continue processing.
- Restart and resume preserve committed work.
- Concurrent starts cannot create two active updates.
- An update never reports completion while batches remain.
- The API exposes progress, products, history, and issues.
- The README documents assumptions, run/test steps, and limitations.
