package httperr

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorResponse is the JSON body for client-facing errors ({ "error": "..." }).
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteError writes a JSON error envelope with the given HTTP status.
func WriteError(c *gin.Context, status int, msg string) {
	c.JSON(status, ErrorResponse{Error: msg})
}

// FirstValidationError returns a short message for the first validation error.
func FirstValidationError(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "validation failed"
	}
	e := errs[0]
	return e.Field() + ": " + tagMessage(e.Tag())
}

func tagMessage(tag string) string {
	switch tag {
	case "required":
		return "required"
	case "email":
		return "must be a valid email"
	case "min":
		return "too short"
	default:
		return "invalid"
	}
}
