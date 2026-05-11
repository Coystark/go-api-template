package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
// @Failure			400		{object}	ErrorResponse
// @Failure			401		{object}	ErrorResponse
// @Failure			500		{object}	ErrorResponse
// @Router			/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var in LoginDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}

	out, err := h.svc.Login(c.Request.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
		default:
			var valErr validator.ValidationErrors
			if errors.As(err, &valErr) {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: firstValidationError(valErr)})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, out)
}

func firstValidationError(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "validation failed"
	}
	e := errs[0]
	return e.Field() + ": invalid"
}
