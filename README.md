# Supplier Catalog Updates

A local-first application that helps staff ingest supplier data batches, track
the progress of catalog updates, review changes and issues, and resume an
update that was interrupted. The backend is written in Go and stores data in
SQLite; a React/Vite dashboard is provided as an optional interface. The demo
uses only fictional supplier data and needs no accounts or real supplier
connections.

**Live demo:** <https://coderpush-api.onrender.com>

## Contents

- [Features](#features)
- [Requirements](#requirements)
- [Quick start](#quick-start)
- [Running the dashboard](#running-the-dashboard)
- [Sample data and demo](#sample-data-and-demo)
- [Deployment](#deployment)
- [Testing](#testing)
- [API](#api)
- [Processing rules and assumptions](#processing-rules-and-assumptions)
- [Limitations and future work](#limitations-and-future-work)
- [Submission links](#submission-links)
- [Use of AI tools](#use-of-ai-tools)
- [Repository layout](#repository-layout)

## Features

- **Update tracking:** start an update, receive product batches, and inspect
  its status, changes, per-batch summaries, and open issues.
- **Full product lifecycle:** new products, changes to existing products, and
  withdrawals. Withdrawn products are kept as tombstones and never physically
  deleted.
- **Safe replays and ordering:** replaying the same batch never creates
  duplicate products. Revisions older than the stored one are ignored, so stale
  data can never overwrite newer data or revive a withdrawn product.
- **Issue tracking:** invalid data and supplier outages are recorded as issues.
  An update cannot be completed while any issue is unresolved.
- **Durable progress:** the catalog and update progress are persisted in
  SQLite. On restart, an in-progress update is marked `interrupted` and can be
  resumed; committed batch work is preserved.
- **Interfaces:** an HTTP JSON API and a web dashboard. No network calls are
  made to real suppliers.

## Requirements

- Go 1.22 or later.
- Node.js 22 or later and npm, only if you run the dashboard in development
  mode or build the UI.
- Docker, only if you want to build the container image.

## Quick start

From the repository root:

```sh
go run ./cmd/catalogd -addr :8080 -data ./data/catalog.db
```

The application creates the database directory, file, and SQLite schema
automatically. Open <http://localhost:8080/health> to check that the service
is running. Stop it with `Ctrl+C`; data is kept in `./data/catalog.db` across
runs.

### Command-line flags

| Flag | Default | Description |
|---|---|---|
| `-addr` | `:8080`, or `:$PORT` when `PORT` is set | HTTP listen address |
| `-data` | `./data/catalog.db` | Path to the SQLite database |
| `-seed` | _(none)_ | Import products from a SQLite mock database, then exit |
| `-static` | _(none)_ | Directory of built dashboard files to serve at `/` |

## Running the dashboard

### Development mode

Start the backend in one terminal as shown above. In a second terminal:

```sh
cd web
npm ci
npm run dev
```

Open the URL printed by Vite (usually <http://localhost:5173>). Vite proxies
API requests to the backend at `http://localhost:8080`; edit
`web/vite.config.ts` to change the backend address.

### Production build

Build the dashboard and let the Go server serve it alongside the API:

```sh
cd web
npm ci
npm run build
cd ..
go run ./cmd/catalogd -addr :8080 -data ./data/catalog.db -static ./web/dist
```

Then open <http://localhost:8080>.

### Entering products

The **Submit a batch** form lets staff enter products row by row (prices in
dollars), fill in an example, or import a CSV file. The CSV needs a header row
with at least `id` and `revision`; the full set of columns is
`id,name,description,price,revision,withdrawn` (`price_cents` is also accepted).
A template can be downloaded from the form.

## Sample data and demo

`examples/mock-data.db` contains 100 fictional products. Import them into the
local database with:

```sh
go run ./cmd/catalogd -data ./data/catalog.db -seed ./examples/mock-data.db
```

The seed command exits once the import finishes. It is safe to run repeatedly:
matching products are skipped, while an existing ID with different content
makes the import fail instead of overwriting data. Afterwards, start the server
without `-seed`.

To run the end-to-end walkthrough against a running server:

```sh
curl -X POST http://localhost:8080/demo
curl http://localhost:8080/updates/demo-spring-2026
curl http://localhost:8080/catalog
curl 'http://localhost:8080/catalog?include_withdrawn=true'
```

The demo builds a baseline catalog and then ingests a spring update containing
changed products, new products, a withdrawn product, and a stale revision that
arrives late. The endpoint is safe to call again: it uses fixed IDs and never
creates duplicate products. After calling it, the results are also visible in
the dashboard.

## Deployment

### Docker

The `Dockerfile` builds the dashboard and the Go binary into a single small
image. On startup the container seeds the sample data (idempotently) and then
serves both the API and the dashboard:

```sh
docker build -t supplier-catalog .
docker run --rm -p 8080:8080 supplier-catalog
```

Open <http://localhost:8080>. The container listens on `$PORT` when set and
stores its database at `$DATA_PATH` (default `/app/data/catalog.db`).

### Render

`render.yaml` defines a Docker-based web service with `/health` as its health
check. To deploy:

1. Push the repository to GitHub.
2. In the Render dashboard, choose **New → Blueprint** and select the
   repository.
3. Render builds the image, starts the service, and redeploys on every push to
   the default branch.

> **Note:** the free plan's filesystem is ephemeral. The SQLite database is
> reset to the 100 sample products on every deploy or restart. To keep data,
> use a paid plan and attach a persistent disk mounted at `/app/data`.

## Testing

Backend tests:

```sh
go test ./...
```

Dashboard tests:

```sh
cd web
npm ci
npm test
```

Type-check and bundle the dashboard:

```sh
cd web
npm ci
npm run build
```

## API

All requests and responses use JSON.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/health` | Service health check |
| `GET` | `/catalog` | List active products |
| `GET` | `/catalog?include_withdrawn=true` | Include withdrawn products |
| `GET` | `/updates` | List updates and their status |
| `POST` | `/updates` | Start an update, e.g. `{"id":"spring-2026"}`; send `{}` to generate an ID |
| `GET` | `/updates/{id}` | Update report: status, batch summaries, changes, and issues |
| `POST` | `/updates/{id}/batches` | Submit a product batch |
| `POST` | `/updates/{id}/resume` | Resume an interrupted update |
| `POST` | `/updates/{id}/failures` | Record a supplier outage and interrupt the update |
| `POST` | `/updates/{id}/issues/{issueID}/resolve` | Mark an issue as resolved |
| `POST` | `/updates/{id}/complete` | Signal that the supplier has delivered all batches |
| `POST` | `/demo` | Run the fictional walkthrough |

`include_withdrawn` accepts a single boolean value (`true`, `false`, `1`,
`0`, …); anything else returns `400 Bad Request`. Errors use the shape
`{"error": "message"}`. See [`docs/API_SPEC.md`](docs/API_SPEC.md) for the
full specification.

### Example: start an update and submit a batch

```sh
curl -X POST http://localhost:8080/updates \
  -H 'Content-Type: application/json' \
  -d '{"id":"spring-2026"}'

curl -X POST http://localhost:8080/updates/spring-2026/batches \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "batch-001",
    "products": [
      {"id": "TEA-100", "name": "Harbor Breakfast Tea", "description": "Black tea", "price_cents": 1399, "revision": 12},
      {"id": "MUG-200", "revision": 8, "withdrawn": true}
    ]
  }'

curl http://localhost:8080/updates/spring-2026
```

### Update reports

An update report contains:

- `status`: one of `in_progress`, `interrupted`, or `completed`.
- `batches`: per-batch counts of applied, stale, unchanged, and rejected
  products, plus the changes each batch made.
- `changes`: the list of catalog changes.
- `issues`: data and outage issues, with their resolution state.

An update completes only when the client calls `/complete` and no issues remain
open. The server cannot know whether the supplier still has batches to send, so
the supplier adapter or staff member must call `/complete` only after the
supplier signals that delivery is finished.

## Processing rules and assumptions

- Suppliers provide stable product IDs, stable batch IDs, and a positive
  integer revision that increases per product. Revisions determine ordering;
  the time a batch is received does not decide which data is newer.
- A revision lower than the stored revision is reported as stale and leaves
  the catalog unchanged. An equal revision with identical content is
  unchanged; an equal revision with a different payload raises a conflict
  issue. Only a newer revision can change a product's state, including
  withdrawing or reactivating it.
- Replaying a batch with the same ID and content is idempotent. Reusing a batch
  ID with different content is rejected and recorded as an issue. Batches are
  processed synchronously: catalog changes, summaries, and issues are persisted
  together before the request succeeds.
- An update that is in progress when the process restarts becomes
  `interrupted`. Staff call `resume` and then continue sending batches. If it is
  unclear whether a batch was stored before a response was lost, resend it with
  the same batch ID and payload.
- Only one unfinished update can exist at a time within a process. Update IDs
  cannot be reused.
- Invalid product data raises an issue while the other valid products in the
  batch are still processed. Open issues, including outage issues, must be
  resolved by staff before the update can complete.
- A product missing from a delivery does not mean it was withdrawn; the
  supplier must send an explicit `withdrawn` event with a revision.

Design rationale is documented in [`docs/DECISIONS.md`](docs/DECISIONS.md).

## Limitations and future work

- This is a demo, local-first application. It has no authentication,
  authorization, audit identity, rate limiting, or security hardening for
  exposure to the internet.
- There is no real supplier adapter, per-supplier retry/backoff,
  supplier-specific schema, or reconciliation of missing batches.
- SQLite and the current update lock are designed for a single instance;
  exclusive updates are not guaranteed across multiple instances. Receipts,
  history, and issues should be normalized into their own tables before
  supporting advanced queries or multi-instance workers.
- Reports summarize results per batch and do not durably store the outcome of
  each individual input product. There is no pagination, advanced filtering,
  or operational metrics yet.

## Submission links

The assessment accepts either a 3–5 minute video or a URL to a deployed
application.

- Video: _not yet available_
- Deployed application: <https://coderpush-api.onrender.com> (free Render
  instance; the first request after idling may take up to a minute while it
  wakes up, and data resets to the sample catalog on restart)

## Use of AI tools

AI tools were used to help scaffold and review the implementation, tests, and
documentation. The author has reviewed the design decisions, source code, and
assumptions and is prepared to explain them in the interview.

## Repository layout

| Path | Contents |
|---|---|
| `cmd/catalogd/` | Executable entry point and command-line flags |
| `internal/catalog/domain/` | Domain models and errors |
| `internal/catalog/service/` | Use cases, update policy, and persistence orchestration |
| `internal/catalog/repository/sqlite/` | SQLite schema, migrations, and storage |
| `internal/catalog/seed/` | Mock-database import |
| `internal/catalog/httpapi/` | JSON HTTP API and demo endpoint |
| `web/` | React/Vite dashboard |
| `examples/` | Fictional sample data |
| `docs/` | Specifications and design documents |
| `Dockerfile`, `render.yaml` | Container image and Render deployment |
