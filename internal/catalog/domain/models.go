package domain

import "time"

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	PriceCents  int64  `json:"price_cents"`
	Revision    int64  `json:"revision"`
	Withdrawn   bool   `json:"withdrawn"`
}

type Change struct {
	ProductID string   `json:"product_id"`
	Kind      string   `json:"kind"`
	Before    *Product `json:"before,omitempty"`
	After     Product  `json:"after"`
}

type Issue struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	BatchID   string    `json:"batch_id,omitempty"`
	ProductID string    `json:"product_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Resolved  bool      `json:"resolved"`
}

type BatchSummary struct {
	ID        string    `json:"id"`
	Digest    string    `json:"digest"`
	Received  time.Time `json:"received_at"`
	Applied   int       `json:"applied"`
	Stale     int       `json:"stale"`
	Unchanged int       `json:"unchanged"`
	Rejected  int       `json:"rejected"`
	Changes   []Change  `json:"changes"`
}

type Update struct {
	ID          string                  `json:"id"`
	Status      string                  `json:"status"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	CompletedAt *time.Time              `json:"completed_at,omitempty"`
	Batches     map[string]BatchSummary `json:"batches"`
	Changes     []Change                `json:"changes"`
	Issues      []Issue                 `json:"issues"`
}

type IncomingProduct struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Revision    int64  `json:"revision"`
	Withdrawn   bool   `json:"withdrawn"`
}

type IncomingBatch struct {
	ID       string            `json:"id"`
	Products []IncomingProduct `json:"products"`
}

type State struct {
	Catalog map[string]Product `json:"catalog"`
	Updates map[string]*Update `json:"updates"`
}
