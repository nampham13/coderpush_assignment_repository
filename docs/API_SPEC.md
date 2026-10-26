# HTTP API Specification

## Conventions

Base path `/api/v1`; JSON request/response; UTC RFC3339 timestamps.
Error shape:

``` json
{"error":{"code":"UPDATE_ALREADY_ACTIVE","message":"An update is already active."}}
```

Never return stack traces. Limit body size. Parse JSON strictly. Do not
accept arbitrary file paths from clients.

## Endpoints

-   `GET /healthz` -\> `200 {"status":"ok"}`.
-   `POST /api/v1/updates` body `{"fixture_id":"demo-basic"}` -\> `201`
    with persisted update summary. Unknown fixture -\> 404; active
    update exists -\> 409. Snapshot fixture and create batches before
    responding.
-   `GET /api/v1/updates?limit=20&offset=0` -\> paginated summaries;
    default limit 20, max 100.
-   `GET /api/v1/updates/{id}` -\> update details, batch status counts,
    total/processed event counts and outcome counts.
-   `POST /api/v1/updates/{id}/resume` -\> `202` if resumable; completed
    update -\> `409 UPDATE_NOT_RESUMABLE`. Repeated calls must not start
    duplicate workers.
-   `POST /api/v1/updates/{id}/batches/{batchID}/retry` -\> `202` only
    for retryable batch. Validate batch belongs to update.
-   `GET /api/v1/products?status=ACTIVE&limit=20&offset=0` -\> paginated
    products.
-   `GET /api/v1/products/{productID}` -\> product or 404.
-   `GET /api/v1/products/{productID}/history?limit=50&offset=0` -\>
    stable ordered history.
-   `GET /api/v1/issues?update_id=...&status=OPEN&limit=50&offset=0` -\>
    paginated issues.
-   `POST /api/v1/issues/{issueID}/resolve` -\> mark resolved; resolving
    twice is idempotent.

## HTTP status/error codes

-   400 `INVALID_REQUEST`: malformed JSON/query/path parameter.
-   404 `UPDATE_NOT_FOUND`, `PRODUCT_NOT_FOUND`, `FIXTURE_NOT_FOUND`,
    `ISSUE_NOT_FOUND`.
-   409 `UPDATE_ALREADY_ACTIVE`, `UPDATE_NOT_RESUMABLE`,
    `BATCH_NOT_RETRYABLE`.
-   422 `VALIDATION_FAILED`.
-   500 `INTERNAL_ERROR`.
-   503 `DATABASE_UNAVAILABLE`.

## API rules

-   Pagination values must be validated; limit must be 1..100, offset
    \>= 0.
-   Stable sort order required for paginated responses.
-   Progress is based on persistent state, never in-memory-only
    counters.
-   Retry/resume request is idempotent with respect to side effects.
-   A request cannot retry a completed batch or resume a completed
    update.
-   MVP may run a sequential background worker after create; document
    lifecycle and graceful shutdown in README.
