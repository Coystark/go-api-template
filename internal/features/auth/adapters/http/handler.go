package authhttp

import (
	"net/http"

	authapp "github.com/caiohenrique/go-api-template/internal/features/auth/app"
	"github.com/caiohenrique/go-api-template/internal/platform/ginbind"
	"github.com/caiohenrique/go-api-template/internal/platform/httperr"
	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP endpoints for authentication.
type Handler struct {
	svc *authapp.Service
}

// NewHandler creates the auth HTTP handler.
func NewHandler(svc *authapp.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers auth routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/auth/login", h.Login)
}

// @Summary			Login
// @Description		Autentica credenciais e retorna JWT.
// @Tags			auth
// @Accept			json
// @Produce			json
// @Param			body	body		authapp.LoginInput	true	"Credenciais"
// @Success			200		{object}	authapp.TokenView
// @Failure			400		{object}	httperr.ErrorResponse
// @Failure			401		{object}	httperr.ErrorResponse
// @Failure			500		{object}	httperr.ErrorResponse
// @Router			/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	in, ok := ginbind.JSON[authapp.LoginInput](c, "invalid json body")
	if !ok {
		return
	}

	out, err := h.svc.Login(c.Request.Context(), in)
	if err != nil {
		httperr.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}
