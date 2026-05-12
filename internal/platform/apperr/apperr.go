package apperr

import "net/http"

type AppError struct {
	Status  int
	Message string
}

func (e *AppError) Error() string { return e.Message }

func NotFound(msg string) *AppError { return &AppError{http.StatusNotFound, msg} }

func Conflict(msg string) *AppError { return &AppError{http.StatusConflict, msg} }

func Unauthorized(msg string) *AppError { return &AppError{http.StatusUnauthorized, msg} }

func BadRequest(msg string) *AppError { return &AppError{http.StatusBadRequest, msg} }
