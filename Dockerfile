# Build the React dashboard.
FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build the Go service (modernc.org/sqlite is pure Go, so CGO is not needed).
FROM golang:1.23-alpine AS api
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/catalogd ./cmd/catalogd

FROM alpine:3.20
WORKDIR /app
COPY --from=api /out/catalogd ./catalogd
COPY --from=web /src/web/dist ./web
COPY examples/mock-data.db ./examples/mock-data.db
ENV DATA_PATH=/app/data/catalog.db
EXPOSE 8080
# Seeding is idempotent, so it is safe on every start (and needed on Render's ephemeral disk).
CMD ["sh", "-c", "mkdir -p \"$(dirname \"$DATA_PATH\")\" && ./catalogd -data \"$DATA_PATH\" -seed ./examples/mock-data.db && exec ./catalogd -data \"$DATA_PATH\" -static ./web"]
