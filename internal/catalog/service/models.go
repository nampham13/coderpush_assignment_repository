package service

import (
	"sync"

	"example.com/supplier-catalog/internal/catalog/domain"
	"example.com/supplier-catalog/internal/catalog/repository/sqlite"
)

type Product = domain.Product
type Change = domain.Change
type Issue = domain.Issue
type BatchSummary = domain.BatchSummary
type Update = domain.Update
type incomingProduct = domain.IncomingProduct
type incomingBatch = domain.IncomingBatch
type state = domain.State

var (
	errNotFound      = domain.ErrNotFound
	errBatchConflict = domain.ErrBatchConflict
	errConflict      = domain.ErrConflict
	errInvalid       = domain.ErrInvalid
)

type Store struct {
	mu   sync.RWMutex
	repo *sqlite.Repository
	data state
}
