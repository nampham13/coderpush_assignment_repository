package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrBatchConflict = errors.New("batch ID conflict")
	ErrConflict      = errors.New("conflict")
	ErrInvalid       = errors.New("invalid request")
)
