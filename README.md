# Supplier Catalog Updates

A local-first application that helps staff ingest supplier data batches, track update progress, review catalog changes and issues, and resume an interrupted update. The backend is written in Go and stores data in SQLite; a React/Vite dashboard is included. The demo uses fictional supplier data and does not require a real supplier account.

**Live demo:** <https://coderpush-api.onrender.com>

## Features

- Start an update, submit product batches, and review status, changes, per-batch summaries, and issues.
- Add products, update product details, or withdraw products. Withdrawn products are retained as tombstones rather than deleted.
- Safely replay batches. Older product revisions are ignored and cannot overwrite newer data or restore a withdrawn product.
- Persist catalog and update progress in SQLite. After a restart, an in-progress update is marked interrupted and can be resumed.
- Record invalid product data and supplier outages as issues. An update cannot complete while unresolved issues remain.
- Use the HTTP JSON API or the web dashboard. The application does not connect to a real supplier.

## Requirements

- Go 1.22 or later.
- Node.js 22 or later and npm, to run or build the dashboard.
- Docker, only for the container instructions below.

## Run locally

From the repository root, start the backend:

```sh
go run ./cmd/catalogd -addr :8080 -data ./data/catalog.db
```

The application creates the database directory and SQLite schema automatically. Check that the backend is running at <http://localhost:8080/health>. Data is stored in `./data/catalog.db` and persists across restarts.

The `-addr` flag sets the listen address and `-data` sets the SQLite database path. If `PORT` is set, it is used as the default port.

## Run the dashboard

Start the backend as above. In another terminal:

```sh
cd web
npm ci
npm run dev
```

Open the URL printed by Vite, usually <http://localhost:5173>. Vite proxies API requests to `http://localhost:8080`.

The dashboard's **Submit a batch** form supports entering products, loading an example, or importing CSV. CSV headers include `id` and `revision`; optional columns are `name`, `description`, `price`, and `withdrawn`.

To build the dashboard and serve it from the Go application:

```sh
cd web
npm ci
npm run build
cd ..
go run ./cmd/catalogd -addr :8080 -data ./data/catalog.db -static ./web/dist
```

Then open <http://localhost:8080>.

## Run tests

Backend tests (from the repository root):

```sh
go test ./...
```

Dashboard tests:

```sh
cd web
npm ci
npm test
```

The Go tests cover the core update rules:

- Replaying a batch is idempotent, older revisions are ignored, and withdrawn products are not restored by stale data.
- Reusing a batch ID with different content is rejected and recorded as an issue.
- After a restart, already applied batches are preserved, the update is marked interrupted, and it is not reported as complete until resumed and explicitly completed.
- Rejected products and supplier outages block completion until their issues are resolved.
- Importing the sample database twice does not create duplicates.
- HTTP routes, query validation, and the demo endpoint.

The dashboard tests cover CSV parsing, the batch form, and the products table.

## Sample data and demo

`examples/mock-data.db` contains 100 fictional products. Import them into the local catalog with:

```sh
go run ./cmd/catalogd -data ./data/catalog.db -seed ./examples/mock-data.db
```

The seed command imports the data and exits. It can be run repeatedly: matching products are skipped, and conflicting existing product IDs are not overwritten. Start the server without `-seed` after importing.

To run the fictional end-to-end walkthrough against a running backend:

```sh
curl -X POST http://localhost:8080/demo
curl http://localhost:8080/updates/demo-spring-2026
curl http://localhost:8080/catalog
curl 'http://localhost:8080/catalog?include_withdrawn=true'
```

The demo creates a baseline catalog, then applies batches containing a changed product, a new product, a withdrawal, and an older revision arriving late. The endpoint uses fixed IDs and is safe to call again. Results are available in the dashboard and update report.

## Deployment

### Docker

Build and run the container:

```sh
docker build -t supplier-catalog .
docker run --rm -p 8080:8080 supplier-catalog
```

Open <http://localhost:8080>. The container serves the API and dashboard. It seeds fictional sample products on startup. The database path defaults to `/app/data/catalog.db` and can be changed with `DATA_PATH`.

### Render

`render.yaml` defines a Docker-based web service with `/health` as its health check. Push the repository to GitHub, then create a Blueprint in Render and select the repository. Render builds and deploys the service.

The free Render filesystem is ephemeral, so the database may reset to the sample catalog after a restart or deploy. Use a persistent disk mounted at `/app/data` to retain data.

## API

All requests and responses use JSON.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/health` | Service health check |
| `GET` | `/catalog` | List active products |
| `GET` | `/catalog?include_withdrawn=true` | Include withdrawn products |
| `GET` | `/updates` | List updates and status |
| `POST` | `/updates` | Start an update; send `{"id":"spring-2026"}` or `{}` for a generated ID |
| `GET` | `/updates/{id}` | View update status, batch summaries, changes, and issues |
| `POST` | `/updates/{id}/batches` | Submit a product batch |
| `POST` | `/updates/{id}/resume` | Resume an interrupted update |
| `POST` | `/updates/{id}/failures` | Record a supplier outage |
| `POST` | `/updates/{id}/issues/{issueID}/resolve` | Resolve an issue |
| `POST` | `/updates/{id}/complete` | Signal that the supplier has delivered all batches |
| `POST` | `/demo` | Run the fictional walkthrough |

Example: create an update, submit a batch, and read its report:

```sh
curl -X POST http://localhost:8080/updates \
  -H 'Content-Type: application/json' \
  -d '{"id":"spring-2026"}'

curl -X POST http://localhost:8080/updates/spring-2026/batches \
  -H 'Content-Type: application/json' \
  -d '{
    "id":"batch-001",
    "products":[
      {"id":"TEA-100","name":"Harbor Breakfast Tea","description":"Black tea","price_cents":1399,"revision":12},
      {"id":"MUG-200","revision":8,"withdrawn":true}
    ]
  }'

curl http://localhost:8080/updates/spring-2026
```

An update report contains its status, per-batch counts and changes, overall catalog changes, and issues. An update remains `in_progress` until `/complete` is called. Completion requires the supplier's completion signal and no unresolved issues; the application cannot infer whether more batches are still expected.

## Assumptions and limitations

- The supplier provides stable product IDs and batch IDs, plus a positive revision that increases per product. Revisions, rather than arrival time, determine which product data is newer.
- A lower revision is ignored. Equal revisions with matching data are unchanged; conflicting data at the same revision creates an issue. A higher revision can update or withdraw a product.
- Replaying a batch with the same ID and content is idempotent. Reusing a batch ID with different content is rejected and recorded as an issue.
- Batch changes, summaries, and issues are persisted together. If a response is lost, resubmit the same batch ID and content safely.
- Only one unfinished update can exist at a time in a process. An interrupted update must be resumed before another update can start. Update IDs cannot be reused.
- A product missing from a batch is not considered withdrawn; the supplier must send an explicit withdrawal with a revision.
- Batches reach the application through the API or dashboard; this stands in for the supplier delivering them.

## Unfinished work

- **No simulated supplier feed.** Staff or a script submit batches, and outages are recorded manually. A fuller version would pull pages from a fake supplier with fault injection, save a cursor after each committed batch, and continue from that cursor on resume. Today, resume reopens the update, and the remaining batches must still be submitted. Already applied batches are kept, and replaying them is safe.
- **Completion relies on an explicit signal.** The supplier does not say how many batches to expect, so the update is completed by calling `/complete`. A batch count or last-batch marker from the supplier would let the application verify this itself.
- **Every change rewrites the full state.** Each change copies the whole state and rewrites it in one SQLite transaction. This is simple and atomic, but slow for large catalogs. Row-level upserts would scale better.
- **Single instance, no authentication.** There is one process, an in-memory lock, and no user accounts or audit of who made each change.
- **No dedicated tests for the SQLite repository.** It is covered indirectly by the restart tests in the service package.

## Assessment submission

- **Demo application:** <https://coderpush-api.onrender.com>
- **Video:** not provided; the deployed application is supplied for review.

## Use of AI tools

The design is my own: I researched the problem and decided the update rules (revision-based ordering, idempotent batch IDs, withdrawn products kept as tombstones, interruption on restart, and completion blocked by unresolved issues), along with the overall structure. AI coding assistants wrote most of the implementation code and tests from that design, and helped draft this README. I reviewed and tested the generated code and can explain or change any part of it.

## Repository layout

- `cmd/catalogd/`: Go application entry point.
- `internal/catalog/domain/`: catalog models and domain errors.
- `internal/catalog/service/`: update rules and use cases.
- `internal/catalog/repository/sqlite/`: SQLite persistence.
- `internal/catalog/seed/`: import fictional sample data.
- `internal/catalog/httpapi/`: HTTP API and demo endpoint.
- `web/`: React/Vite dashboard.
- `examples/`: fictional sample database.
- `Dockerfile`, `render.yaml`: container build and Render deployment configuration.
