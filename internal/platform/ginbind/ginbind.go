package ginbind

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the JSON envelope used when binding fails ({ "error": "..." }).
type ErrorResponse struct {
	Error string `json:"error"`
}

// JSON decodes the request body into T using Gin's binding (JSON and validation tags).
// On failure it writes 400 with badRequestMsg and returns (zero value of T, false).
func JSON[T any](c *gin.Context, badRequestMsg string) (T, bool) {
	var v T
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: badRequestMsg})
		return v, false
	}
	return v, true
}
