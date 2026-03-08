package httpapi

import (
	"errors"
	"net/http"

	"charon/backend/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorEnvelope struct {
	ErrorCode   string       `json:"error_code"`
	Message     string       `json:"message"`
	TraceID     string       `json:"trace_id"`
	FieldErrors []fieldError `json:"field_errors,omitempty"`
}

func respondValidationError(c *gin.Context, message string, fieldErrors ...fieldError) {
	c.JSON(http.StatusBadRequest, errorEnvelope{
		ErrorCode:   "VALIDATION_ERROR",
		Message:     message,
		TraceID:     traceIDFromContext(c),
		FieldErrors: fieldErrors,
	})
}

func respondError(c *gin.Context, status int, code string, message string, fieldErrors ...fieldError) {
	c.JSON(status, errorEnvelope{
		ErrorCode:   code,
		Message:     message,
		TraceID:     traceIDFromContext(c),
		FieldErrors: fieldErrors,
	})
}

func respondInternalError(c *gin.Context) {
	respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
}

func respondAuthError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, auth.ErrInvalidCredentials),
		errors.Is(err, auth.ErrRefreshTokenInvalid),
		errors.Is(err, auth.ErrRefreshTokenExpired),
		errors.Is(err, auth.ErrAccessTokenInvalid),
		errors.Is(err, auth.ErrAccessTokenExpired),
		errors.Is(err, auth.ErrAuthorizationRequired):
		status = http.StatusUnauthorized
	case errors.Is(err, auth.ErrAccountDisabled),
		errors.Is(err, auth.ErrInsufficientRole):
		status = http.StatusForbidden
	}

	respondError(c, status, auth.ErrorCode(err), auth.ErrorMessage(err))
}
