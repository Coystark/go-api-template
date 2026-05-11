package ginbind

import (
	"net/http"

	"github.com/caiohenrique/go-api-template/internal/platform/httperr"
	"github.com/caiohenrique/go-api-template/internal/platform/pagination"
	"github.com/gin-gonic/gin"
)

// JSON decodes the request body into T using Gin's binding (JSON and validation tags).
// On failure it writes 400 with badRequestMsg and returns (zero value of T, false).
func JSON[T any](c *gin.Context, badRequestMsg string) (T, bool) {
	var v T
	if err := c.ShouldBindJSON(&v); err != nil {
		httperr.WriteError(c, http.StatusBadRequest, badRequestMsg)
		return v, false
	}
	return v, true
}

// Query decodes query parameters into T using Gin's binding.
// On failure it writes 400 with badRequestMsg and returns (zero value of T, false).
func Query[T any](c *gin.Context, badRequestMsg string) (T, bool) {
	var v T
	if err := c.ShouldBindQuery(&v); err != nil {
		httperr.WriteError(c, http.StatusBadRequest, badRequestMsg)
		return v, false
	}
	return v, true
}

// Pagination lê os parâmetros de paginação padrão da query string.
func Pagination(c *gin.Context) (pagination.Query, bool) {
	query, ok := Query[pagination.Query](c, "invalid pagination query")
	if !ok {
		return pagination.Query{}, false
	}
	return pagination.Normalize(query), true
}
