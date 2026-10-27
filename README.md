# Supplier catalog updates

A small, local-first Go HTTP service for applying fictional supplier catalog batches. It stores catalog products and update records in SQLite, so data survives application restarts without requiring a live supplier account.

## Project layout

- `cmd/catalogd/` contains the executable entry point.
- `internal/catalog/domain/` contains shared catalog models and domain errors.
- `internal/catalog/service/` contains catalog use cases, update rules, and tests.
- `internal/catalog/repository/sqlite/` contains SQLite migrations and storage.
- `internal/catalog/seed/` contains the mock-database import use case and tests.
- `internal/catalog/httpapi/` contains HTTP handlers and API tests.
- `web/` contains the React dashboard for catalog and update operations.
- `docs/` contains the product and architecture specifications.
- `examples/` contains fictional input data.

## Run

Requires Go 1.22 or later.

```sh
go run ./cmd/catalogd -addr :8080 -data ./data/catalog.db
```

The SQLite database and schema are created automatically. Run the tests with:

```sh
go test ./...
```

## Seed mock data

The repository includes 100 fictional products and demo events in the SQLite
fixture [`examples/mock-data.db`](./examples/mock-data.db). Seed the products into
the default local database with:

```sh
go run ./cmd/catalogd -data ./data/catalog.db -seed ./examples/mock-data.db
```

The command is idempotent: running it again leaves matching products unchanged
and reports `seeded 0 products`. An existing product ID with different data
causes the seed to fail rather than overwrite catalog data. Start the server
after seeding with the `-seed` option omitted.

## Run the web interface

Start the backend in one terminal:

```sh
go run ./cmd/catalogd -addr :8080 -data ./data/catalog.db
```

In another terminal, install the web dependencies and start the development
server:

```sh
cd web
npm install
npm run dev
```

Open the URL printed by Vite. The development server proxies API requests to
`http://localhost:8080`. Run the interface checks with `npm test` and create a
production bundle with `npm run build`.

## Try the demo

In a second terminal, run:

```sh
curl -X POST http://localhost:8080/demo
curl http://localhost:8080/updates/demo-spring-2026
curl http://localhost:8080/catalog
curl 'http://localhost:8080/catalog?include_withdrawn=true'
```

The fictional demo creates a baseline catalog, then applies a seasonal update containing a changed tea, a withdrawn mug, a new coffee, and a deliberately late/older tea revision. Repeating the demo request is safe.

## API

All request and response bodies are JSON. Mutating endpoints persist their changes before returning success.

| Method and path | Purpose |
|---|---|
| `GET /health` | Liveness check |
| `GET /catalog` | List active products |
| `GET /catalog?include_withdrawn=true` | Include withdrawn products |
| `GET /updates` | List update reports |
| `POST /updates` | Start one update; send `{}` for a generated ID or `{"id":"spring-2026"}` for a chosen ID. An existing ID or another unfinished update returns `409` |
| `GET /updates/{id}` | Read status, batch summaries, applied changes, and issues |
| `POST /updates/{id}/batches` | Submit a supplier batch |
| `POST /updates/{id}/resume` | Resume an interrupted update |
| `POST /updates/{id}/failures` | Record an outage and interrupt an update |
| `POST /updates/{id}/issues/{issueID}/resolve` | Mark an issue as addressed |
| `POST /updates/{id}/complete` | Explicitly record the supplier's completion signal |
| `POST /demo` | Seed and run the fictional end-to-end example |

Example batch:

```sh
curl -X POST http://localhost:8080/updates \
  -H 'Content-Type: application/json' \
  -d '{"id":"spring-2026"}'

curl -X POST http://localhost:8080/updates/spring-2026/batches \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "batch-001",
    "products": [
      {"id":"TEA-100","name":"Harbor Breakfast Tea","description":"Black tea","price_cents":1399,"revision":12},
      {"id":"MUG-200","revision":8,"withdrawn":true}
    ]
  }'
```

Create the update first with `POST /updates`. Only one unfinished update may exist at a time; an interrupted update must be resumed or completed before another can start. Update IDs cannot be reused. Batch IDs must be unique within an update. Replaying a batch ID with the same content is an idempotent success; reusing it with different content is rejected and reported as an issue. Product revisions are supplier-provided, positive, and monotonic per product. Lower revisions are recorded as stale and ignored. Equal revisions are unchanged when the content matches and reported as a conflict when it does not.

An update remains `in_progress` until `/complete` is explicitly called. A process restart marks any in-progress update as `interrupted`; completed work is retained and can be continued with `/resume`. Each batch is processed synchronously and its catalog changes, summary, and issues are persisted together before the request succeeds. If interrupted before that save, submit the batch again; if the save succeeded but the response was lost, replaying the same batch is safe. Every unresolved issue, including a supplier outage or rejected product, blocks completion until a staff member addresses it. Product data validation problems are retained as issues while valid records in that batch continue to apply.

## Design choices and assumptions

- Application components are constructed through `StoreFactory` and
  `HandlerFactory`; keep new component creation behind factories.
- Keep each Go source and test file at or below 100 lines. Split implementation
  by cohesive responsibility rather than allowing large files to grow.
- The supplier exposes stable product IDs, stable batch IDs, and a monotonically increasing revision for each product. Without a supplier ordering token, the service cannot reliably distinguish newer from older data.
- `withdrawn: true` is a versioned product event, not a physical deletion. Only a higher supplier revision can change that state.
- "Completed" means the supplier has signaled that all batches for the update have been delivered. The service cannot infer completeness from the batches it has received.
- SQLite stores products in a constrained `products` table and update reports in the `updates` table. Product and update rows are saved in one transaction.
- The current Store serializes state changes in one process. SQLite uses a busy timeout, but update exclusivity is not yet enforced across multiple running application instances.
- The service does not know how many batches the supplier intends to send. Staff or an adapter must use the supplier's completion signal to call `/complete`; the application cannot detect a batch that was never delivered. Batch progress is summarized by batch, without a durable receipt/result row for every input product.
- The application does not make supplier network calls. A staff member or future supplier adapter submits batches through the HTTP API.

## Unfinished work / production considerations

- Normalize update batches, event receipts, product history, and issues into separate SQLite tables before supporting advanced querying or multi-instance workers.
- Add authentication, authorization, audit identity, request rate limits, and deployment configuration before exposing the API beyond a trusted local environment.
- Define a real supplier adapter, supplier-specific schemas, retry/backoff behavior, and reconciliation rules with the supplier.
- Add pagination, advanced filtering, and operational metrics.
- Decide retention policies and how resolved data-quality issues should be corrected or replayed.
- The demo is an API walkthrough, not a deployed application or a video.

## AI tools

AI assistance was used to scaffold and review the implementation, tests, and documentation. The code and assumptions should be independently reviewed before production use.

## Specification documents

The repository also includes a specification pack under `docs/`, an implementation task queue in `TASK_QUEUE.md`, and the SQLite mock fixture in `examples/mock-data.db`. The SQLite schema is intentionally a first persistence increment; normalize remaining update details as the matching tasks are implemented.

The repository-wide Copilot instruction file is `.github/copilot-instruction.md`.
