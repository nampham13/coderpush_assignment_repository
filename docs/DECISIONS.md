# Architecture Decisions

  -----------------------------------------------------------------------
  ID                      Decision                Rationale / trade-off
  ----------------------- ----------------------- -----------------------
  D01                     Go modular monolith     Simple to build and
                                                  demo; not distributed.

  D02                     SQLite                  No separate DB service;
                                                  carefully manage writer
                                                  concurrency.

  D03                     Fictional JSON fixtures No live supplier
                                                  account; does not model
                                                  every network failure.

  D04                     Sequential MVP worker   Correctness and
                                                  recovery before
                                                  throughput.

  D05                     One active update       Simplifies
                                                  orchestration; prevents
                                                  concurrent imports.

  D06                     Per-product supplier    Ordering independent of
                          version                 local clocks; requires
                                                  trustworthy source
                                                  versions.

  D07                     Product ID is           Avoids silently merging
                          case-sensitive          supplier identities.

  D08                     Withdrawal tombstone    Prevents stale data
                                                  from restoring
                                                  withdrawn products.

  D09                     Absence does not imply  Batch may be partial;
                          withdrawal              explicit event
                                                  required.

  D10                     Transaction per event   One bad event does not
                                                  rollback the batch;
                                                  more transactions.

  D11                     Immutable input         Resume processes the
                          snapshot + checksum     same input; uses
                                                  storage.

  D12                     Stable API error codes  Easier client handling
                                                  and tests; codes must
                                                  be maintained.

  D13                     No auth in demo         Focus on assignment;
                          baseline                not safe for public
                                                  production exposure.
  -----------------------------------------------------------------------

## Before production

Clarify supplier version semantics, event ID guarantees, null-field
semantics, full-snapshot behavior, retention, authentication, audit
actor, UI, alerting and parallel worker requirements.
