package usecase

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrInsufficientStock  = errors.New("insufficient available stock")
	ErrCannotCancelDone   = errors.New("cannot cancel a completed transaction")
	ErrStockCannotBeNeg   = errors.New("stock cannot be negative")
	ErrInvalidInput       = errors.New("invalid input")
)
