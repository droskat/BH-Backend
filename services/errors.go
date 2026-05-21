package services

import "errors"

var (
	ErrNotFound  = errors.New("resource not found")
	ErrForbidden = errors.New("access forbidden: ownership mismatch")
)
