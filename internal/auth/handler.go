package auth

import (
	"errors"
	"net/http"

	"github.com/caiohenrique/go-api-template/internal/platform/ginbind"
	"github.com/caiohenrique/go-api-template/internal/platform/httperr"
	"github.com/gin-gonic/gin"
)

// Handler expõe endpoints HTTP de autenticação.
type Handler struct {
	svc *Service
}

// NewHandler cria o handler com dependências injetadas.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registra rotas de auth.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/auth/login", h.Login)
}

// @Summary			Login
// @Description		Autentica credenciais e retorna JWT.
// @Tags			auth
// @Accept			json
// @Produce			json
// @Param			body	body		LoginDTO	true	"Credenciais"
// @Success			200		{object}	TokenResponseDTO
// @Failure			400		{object}	httperr.ErrorResponse
// @Failure			401		{object}	httperr.ErrorResponse
// @Failure			500		{object}	httperr.ErrorResponse
// @Router			/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	in, ok := ginbind.JSON[LoginDTO](c, "invalid json body")
	if !ok {
		return
	}

	out, err := h.svc.Login(c.Request.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			httperr.WriteError(c, http.StatusUnauthorized, "invalid credentials")
		default:
			httperr.WriteValidationOrInternal(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, out)
}
