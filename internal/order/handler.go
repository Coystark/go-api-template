package order

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Handler expõe endpoints HTTP do domínio order.
type Handler struct {
	svc *Service
}

// NewHandler cria o handler com dependências injetadas.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registra rotas sob o router informado.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	rg.POST("/orders", h.Create)
	rg.GET("/orders/:id", h.GetByID)
	rg.PATCH("/orders/:id", h.Update)
}

// @Summary			Criar pedido
// @Description		Cria um pedido com serviços aninhados.
// @Tags			orders
// @Accept			json
// @Produce			json
// @Param			body	body		CreateOrderDTO	true	"Dados do pedido"
// @Success			201		{object}	ResponseDTO
// @Failure			400		{object}	ErrorResponse
// @Failure			500		{object}	ErrorResponse
// @Router			/orders [post]
func (h *Handler) Create(c *gin.Context) {
	var in CreateOrderDTO
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

// @Summary			Buscar pedido
// @Description		Retorna um pedido por ID com seus serviços.
// @Tags			orders
// @Produce			json
// @Param			id	path		string	true	"UUID do pedido"
// @Success			200	{object}	ResponseDTO
// @Failure			400	{object}	ErrorResponse
// @Failure			404	{object}	ErrorResponse
// @Failure			500	{object}	ErrorResponse
// @Router			/orders/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid order id")
		return
	}

	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

// @Summary			Atualizar pedido (parcial)
// @Description		Atualização parcial: título opcional; itens em `order_services` com `id` são UPDATE, sem `id` são INSERT; `removed_service_ids` deleta na mesma transação. Itens omitidos permanecem inalterados.
// @Tags			orders
// @Accept			json
// @Produce			json
// @Param			id		path		string			true	"UUID do pedido"
// @Param			body	body		UpdateOrderDTO	true	"Patch do pedido"
// @Success			200		{object}	ResponseDTO
// @Failure			400		{object}	ErrorResponse
// @Failure			404		{object}	ErrorResponse
// @Failure			500		{object}	ErrorResponse
// @Router			/orders/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid order id")
		return
	}

	var in UpdateOrderDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json body")
		return
	}

	out, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(c, http.StatusNotFound, "order not found")
	case errors.Is(err, ErrServiceNotInOrder):
		writeError(c, http.StatusBadRequest, "order service does not belong to order")
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
	c.JSON(status, ErrorResponse{Error: msg})
}
