# Test Specification

## Layers

1.  Unit: validator, version policy, state aggregation.
2.  Repository integration: SQLite constraints, migrations,
    transactions.
3.  Service: event processing, retry/resume, issue deduplication.
4.  HTTP: handlers via `httptest`.
5.  E2E: fixture -\> update -\> catalog/history/issues/progress.
6.  Recovery: reopen a service using the same temporary database.

Use isolated temporary DB/files per test. No external network or real
supplier credentials.

## Required cases

  -----------------------------------------------------------------------
  ID                      Scenario                Expected
  ----------------------- ----------------------- -----------------------
  U01                     Empty/whitespace        rejected with issue
                          product ID              

  U02                     Invalid version,        rejected with stable
                          negative price/stock,   issue code
                          bad currency/action     

  U03                     Newer version           applied

  U04                     Older version           stale ignored, no
                                                  mutation

  U05                     Same version and same   duplicate/no-op
                          payload                 

  U06                     Same version, different version conflict;
                          payload/action          current state unchanged

  U07                     Withdraw v8 then UPSERT remains withdrawn v8
                          v6                      

  U08                     Reactivate with newer   active if valid
                          version                 

  U09                     Same event ID with      event ID conflict;
                          different payload       original receipt
                                                  preserved

  U10                     Same name, different    two products
                          IDs                     

  R01                     Insert duplicate        unique constraint
                          product ID              

  R02                     History insert fails    transaction rollback
                          after product mutation  

  R03                     Migration runs twice    safe/idempotent

  R04                     Reopen DB               committed state
                                                  persists

  F01                     Crash before event      event is processed on
                          transaction commit      resume

  F02                     Crash after commit but  replay creates no
                          before response         duplicate side effects

  F03                     Crash mid-batch         committed events
                                                  remain; remaining
                                                  events resume

  F04                     Restart with stale      reconciled and
                          RUNNING batch           resumable

  F05                     Snapshot checksum       fail visibly; no silent
                          mismatch                processing

  F06                     Double                  no duplicate active
                          start/resume/retry      update/worker

  H01                     Invalid JSON/query      400, process remains
                                                  alive

  H02                     Unknown IDs             404 with standard error
                                                  schema

  H03                     Resume completed update 409

  H04                     Error response          no stack trace or
                                                  internal filesystem
                                                  path
  -----------------------------------------------------------------------

## E2E scenarios

1.  Duplicate/out-of-order: SKU-1 v1 price 100, v3 price 150, duplicate
    v3, then v2 price 120. Final version 3/price 150, one product, stale
    event recorded.
2.  Withdrawal: SKU-2 v5 UPSERT, v8 WITHDRAW, v6 UPSERT. Final state
    withdrawn v8.
3.  Partial invalid: valid SKU-3, invalid negative-price SKU-4, valid
    SKU-5. SKU-3/SKU-5 apply; issue for SKU-4; update ends
    `COMPLETED_WITH_ISSUES`.
4.  Restart/resume: batch 1 committed, batch 2 pending/running; restart
    preserves batch 1 and resumes remaining work without duplicate side
    effects.
5.  Double start: two concurrent create requests; at most one active
    update.
6.  Empty fixture: behavior matches documented zero-event policy.

## Quality checks

Run `gofmt`, `go test ./...`, `go test -race ./...`, `go vet ./...` when
feasible. Report honestly if a command was not run or unavailable.

## Definition of Done

-   Acceptance criteria met and tests added.
-   No ignored errors or TODOs posing as completed functionality.
-   Test suite passes.
-   README/spec updated for behavior changes.
-   Final report lists changed files, actual commands/results, and known
    limitations.
