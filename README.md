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

## Assessment submission

- **Demo application:** <https://coderpush-api.onrender.com>
- **Video:** not provided; the deployed application is supplied for review.

## Use of AI tools

AI tools were used to assist with implementation and documentation. The author reviewed the code and can explain the design choices and assumptions during the interview.

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
