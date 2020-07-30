package kv

import "errors"

var (
	// ErrAlreadySet refers to create operations that would overwrite an intended value.
	ErrAlreadySet = errors.New("value is already set")
	// ErrInvalidType covers types that are invalid for exchange with this library.
	ErrInvalidType = errors.New("invalid type in k/v operation")
	// ErrUnsetValue is returned when nil is sent because the value is unset.
	ErrUnsetValue = errors.New("unset value")
	// ErrNotEqual indicates that a CAS operation failed its compare.
	ErrNotEqual = errors.New("original value is not equal")
)
