package auth

import (
	"net/http"
	"strings"

	"github.com/caiohenrique/go-api-template/internal/platform/requestctx"
	"github.com/gin-gonic/gin"
)

// NewMiddleware valida Bearer JWT e injeta user ID no contexto da requisição.
func NewMiddleware(tokens *TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}
		id, err := tokens.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Request = c.Request.WithContext(requestctx.WithUserID(c.Request.Context(), id))
		c.Next()
	}
}
