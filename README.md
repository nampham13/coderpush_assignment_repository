# Supplier catalog updates

A small, local-first Go HTTP service for applying fictional supplier catalog batches. It keeps the current catalog and update history in a JSON file, so completed batches survive application restarts without requiring a live supplier account or third-party packages.

## Run

Requires Go 1.22 or later.

```sh
go run . -addr :8080 -data ./data/catalog.json
```

The database file is created automatically. Run the tests with:

```sh
go test ./...
```

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
| `POST /updates` | Start an update; send `{}` for a generated ID or `{"id":"spring-2026"}` for a chosen ID |
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

Create the update first with `POST /updates`. Batch IDs must be unique within an update. Replaying a batch ID with the same content is an idempotent success; reusing it with different content is rejected and reported as an issue. Product revisions are supplier-provided, positive, and monotonic per product. Lower revisions are recorded as stale and ignored. Equal revisions are unchanged when the content matches and reported as a conflict when it does not.

An update remains `in_progress` until `/complete` is explicitly called. A process restart marks any in-progress update as `interrupted`; completed work is retained and can be continued with `/resume`. Every unresolved issue, including a supplier outage or rejected product, blocks completion until a staff member addresses it. Product data validation problems are retained as issues while valid records in that batch continue to apply.

## Design choices and assumptions

- The supplier exposes stable product IDs, stable batch IDs, and a monotonically increasing revision for each product. Without a supplier ordering token, the service cannot reliably distinguish newer from older data.
- `withdrawn: true` is a versioned product event, not a physical deletion. Only a higher supplier revision can change that state.
- "Completed" means the supplier has signaled that all batches for the update have been delivered. The service cannot infer completeness from the batches it has received.
- A single JSON file is suitable for a local demonstration and one application process. Updates are serialized in memory and saved by atomic file replacement; the existing file is not overwritten until the replacement has been written and synced.
- The application does not make supplier network calls. A staff member or future supplier adapter submits batches through the HTTP API.

## Unfinished work / production considerations

- Replace the JSON file with a transactional database for multi-instance use, larger catalogs, stronger recovery guarantees, and concurrent operators.
- Add authentication, authorization, audit identity, request rate limits, and deployment configuration before exposing the API beyond a trusted local environment.
- Define a real supplier adapter, supplier-specific schemas, retry/backoff behavior, and reconciliation rules with the supplier.
- Add pagination, filtering, operational metrics, and a user interface for reviewing and resolving issues.
- Decide retention policies and how resolved data-quality issues should be corrected or replayed.
- The demo is an API walkthrough, not a deployed application or a video.

## AI tools

AI assistance was used to scaffold and review the implementation, tests, and documentation. The code and assumptions should be independently reviewed before production use.
