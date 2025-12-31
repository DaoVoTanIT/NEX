package errors

import "errors"

var (
	ErrNotFound            = errors.New("record not found")
	ErrConflict            = errors.New("record already exists")
	ErrInternalServerError = errors.New("internal server error")
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
)
