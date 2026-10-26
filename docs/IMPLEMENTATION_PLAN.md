# Implementation Plan

Implement one task per Copilot session. Do not ask Copilot to build the
whole product at once.

-   **T-001 Scaffold:** `go.mod`, `cmd/catalogd/main.go`, `.gitignore`,
    minimal README. `go test ./...` works.
-   **T-002 Config/logging:** environment config, startup validation,
    structured logs, no secrets.
-   **T-003 Domain types:** Product, SupplierEvent, Update, Batch,
    Issue, EventOutcome.
-   **T-004 Validator:** required fields and stable validation error
    codes; table-driven tests.
-   **T-005 Version policy:**
    new/stale/duplicate/conflict/withdraw/reactivate; unit tests.
-   **T-006 SQLite/migrations:** DB setup, foreign keys, busy timeout,
    migration tests.
-   **T-007 Product repository:** get/list/insert/conditional update;
    unique product ID tests.
-   **T-008 Update/batch repository:** transactional creation, state
    transitions, single active update.
-   **T-009 Receipts/attempts/history/issues:** dedupe keys, transaction
    consistency, tests.
-   **T-010 Fixture loader:** allowlist, immutable snapshot, checksum,
    deterministic batches.
-   **T-011 Event processor:** validate, dedupe, compare versions,
    atomically persist outcome.
-   **T-012 Batch processor:** claim, per-event processing, final batch
    status, retry policy.
-   **T-013 Orchestration/recovery:** start/resume, reconcile stale
    RUNNING states, status aggregation.
-   **T-014 Update HTTP API:** health, create/list/get/progress.
-   **T-015 Retry/resume API:** safe state validation and idempotency.
-   **T-016 Catalog/issues API:** products, history, issue list/resolve.
-   **T-017 HTTP hardening:** request limits, strict JSON, timeouts,
    graceful shutdown.
-   **T-018 Demo fixtures:** duplicate, stale, withdrawal, invalid
    record, multiple batches.
-   **T-019 E2E tests:** all scenarios in `docs/TEST_SPEC.md`.
-   **T-020 README/demo script:**
    install/run/curl/tests/assumptions/limitations/AI usage.
-   **T-021 Final review:** gofmt, tests, race, vet, secrets scan,
    verify demo from clean checkout.

Dependency order: T-001 -\> T-002 -\> T-003 -\> T-004 -\> T-005 -\>
T-006 -\> T-007/T-008 -\> T-009 -\> T-010 -\> T-011 -\> T-012 -\> T-013
-\> T-014..T-021.

For every task: read specs, inspect existing code, plan briefly,
implement only that task, add tests, run checks, summarize changed files
and real test results.
