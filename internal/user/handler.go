package user

import (
	"errors"
	"net/http"

	"github.com/caiohenrique/go-api-template/internal/platform/requestctx"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

type errorResponse struct {
	Error string `json:"error"`
}

// Create trata POST /users.
func (h *Handler) Create(c *gin.Context) {
	var in CreateDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json body")
		return
	}

	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, out)
}

// Me trata GET /users/me.
func (h *Handler) Me(c *gin.Context) {
	id, ok := requestctx.UserID(c.Request.Context())
	if !ok || id == uuid.Nil {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(c, http.StatusNotFound, "user not found")
	case errors.Is(err, ErrEmailAlreadyExists):
		writeError(c, http.StatusConflict, "email already exists")
	default:
		var valErr validator.ValidationErrors
		if errors.As(err, &valErr) {
			writeError(c, http.StatusBadRequest, firstValidationError(valErr))
			return
		}
		writeError(c, http.StatusInternalServerError, "internal error")
	}
}

func firstValidationError(errs validator.ValidationErrors) string {
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

func writeError(c *gin.Context, status int, msg string) {
	c.JSON(status, errorResponse{Error: msg})
}
