package user

import (
	"net/http"

	"github.com/caiohenrique/go-api-template/internal/platform/ginbind"
	"github.com/caiohenrique/go-api-template/internal/platform/httperr"
	"github.com/caiohenrique/go-api-template/internal/platform/requestctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler expõe endpoints HTTP do domínio user.
type Handler struct {
	svc *Service
}

// NewHandler cria o handler com dependências injetadas.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registra rotas sob o router informado.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	rg.POST("/users", h.Create)
	rg.GET("/users/me", authMiddleware, h.Me)
}

// @Summary			Criar usuário
// @Description		Registra um novo usuário.
// @Tags			users
// @Accept			json
// @Produce			json
// @Param			body	body		CreateDTO	true	"Dados do usuário"
// @Success			201		{object}	ResponseDTO
// @Failure			400		{object}	httperr.ErrorResponse
// @Failure			409		{object}	httperr.ErrorResponse
// @Failure			500		{object}	httperr.ErrorResponse
// @Router			/users [post]
func (h *Handler) Create(c *gin.Context) {
	in, ok := ginbind.JSON[CreateDTO](c, "invalid json body")
	if !ok {
		return
	}

	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httperr.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, out)
}

// @Summary			Usuário autenticado
// @Description		Retorna o perfil do usuário do token JWT.
// @Tags			users
// @Accept			json
// @Produce			json
// @Security		BearerAuth
// @Success			200	{object}	ResponseDTO
// @Failure			401	{object}	httperr.ErrorResponse
// @Failure			404	{object}	httperr.ErrorResponse
// @Failure			500	{object}	httperr.ErrorResponse
// @Router			/users/me [get]
func (h *Handler) Me(c *gin.Context) {
	id, ok := requestctx.UserID(c.Request.Context())
	if !ok || id == uuid.Nil {
		httperr.WriteError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		httperr.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}
