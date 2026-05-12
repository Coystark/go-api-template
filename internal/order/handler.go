package order

import (
	"errors"
	"net/http"

	"github.com/caiohenrique/go-api-template/internal/platform/ginbind"
	"github.com/caiohenrique/go-api-template/internal/platform/httperr"
	"github.com/gin-gonic/gin"
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
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authed := rg.Group("")
	authed.Use(authMiddleware)

	authed.POST("/orders", h.Create)
	authed.GET("/orders", h.List)
	authed.GET("/order-services", h.ListOrderServices)
	authed.GET("/orders/:id", h.GetByID)
	authed.PUT("/orders/:id", h.Update)
	authed.DELETE("/orders/:id", h.Delete)
}

// @Summary			Criar pedido
// @Description		Cria um pedido com serviços aninhados.
// @Tags			orders
// @Accept			json
// @Produce			json
// @Security		BearerAuth
// @Param			body	body		CreateOrderDTO	true	"Dados do pedido"
// @Success			201		{object}	ResponseDTO
// @Failure			400		{object}	httperr.ErrorResponse
// @Failure			401		{object}	httperr.ErrorResponse
// @Failure			500		{object}	httperr.ErrorResponse
// @Router			/orders [post]
func (h *Handler) Create(c *gin.Context) {
	in, ok := ginbind.JSON[CreateOrderDTO](c, "invalid json body")
	if !ok {
		return
	}

	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, out)
}

// @Summary			Listar pedidos
// @Description		Retorna pedidos resumidos com paginação.
// @Tags			orders
// @Produce			json
// @Security		BearerAuth
// @Param			page		query		int		false	"Página, começando em 1"	default(1)
// @Param			page_size	query		int		false	"Itens por página, máximo 100"	default(20)
// @Param			code		query		string	false	"Filtro parcial case-insensitive em `code` (ILIKE)"
// @Param			title		query		string	false	"Filtro parcial case-insensitive em `title` (ILIKE)"
// @Success			200			{object}	ListResponseDTO
// @Failure			400			{object}	httperr.ErrorResponse
// @Failure			401			{object}	httperr.ErrorResponse
// @Failure			500			{object}	httperr.ErrorResponse
// @Router			/orders [get]
func (h *Handler) List(c *gin.Context) {
	query, ok := ginbind.Query[ListQueryDTO](c, "invalid list query")
	if !ok {
		return
	}

	out, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

// @Summary			Listar serviços de pedidos
// @Description		Retorna linhas de order_services com paginação. Filtro opcional `order_id` restringe ao pedido. Apenas serviços de pedidos não removidos (soft delete).
// @Tags			order-services
// @Produce			json
// @Security		BearerAuth
// @Param			page		query		int		false	"Página, começando em 1"	default(1)
// @Param			page_size	query		int		false	"Itens por página, máximo 100"	default(20)
// @Param			order_id	query		string	false	"UUID v7 do pedido (filtro opcional)"
// @Success			200			{object}	ListOrderServicesResponseDTO
// @Failure			400			{object}	httperr.ErrorResponse
// @Failure			401			{object}	httperr.ErrorResponse
// @Failure			500			{object}	httperr.ErrorResponse
// @Router			/order-services [get]
func (h *Handler) ListOrderServices(c *gin.Context) {
	query, ok := ginbind.Query[ListOrderServicesQueryDTO](c, "invalid list order services query")
	if !ok {
		return
	}

	out, err := h.svc.ListOrderServices(c.Request.Context(), query)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

// @Summary			Buscar pedido
// @Description		Retorna um pedido por ID com seus serviços.
// @Tags			orders
// @Produce			json
// @Security		BearerAuth
// @Param			id	path		string	true	"UUID v7 do pedido"
// @Success			200	{object}	ResponseDTO
// @Failure			400	{object}	httperr.ErrorResponse
// @Failure			401	{object}	httperr.ErrorResponse
// @Failure			404	{object}	httperr.ErrorResponse
// @Failure			500	{object}	httperr.ErrorResponse
// @Router			/orders/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httperr.WriteError(c, http.StatusBadRequest, "invalid order id")
		return
	}

	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

// @Summary			Remover pedido
// @Description		Remove logicamente um pedido.
// @Tags			orders
// @Produce			json
// @Security		BearerAuth
// @Param			id	path	string	true	"UUID v7 do pedido"
// @Success			204
// @Failure			400	{object}	httperr.ErrorResponse
// @Failure			401	{object}	httperr.ErrorResponse
// @Failure			404	{object}	httperr.ErrorResponse
// @Failure			500	{object}	httperr.ErrorResponse
// @Router			/orders/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httperr.WriteError(c, http.StatusBadRequest, "invalid order id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary			Atualizar pedido
// @Description		Replace total do cabeçalho: `title` obrigatório; `subject`, `code`, `sent_at` e `converted_at` omitidos ou `null` gravam NULL (reenvie para manter). Em `order_services`, linhas com `id` são UPDATE (sem `id` são INSERT); linhas não enviadas permanecem; `removed_service_ids` remove na mesma transação. Em UPDATE de linha, `end_date` e `observations` omitidos ou `null` gravam NULL — reenvie para mantê-los.
// @Tags			orders
// @Accept			json
// @Produce			json
// @Security		BearerAuth
// @Param			id		path		string			true	"UUID v7 do pedido"
// @Param			body	body		UpdateOrderDTO	true	"Corpo do pedido"
// @Success			200		{object}	ResponseDTO
// @Failure			400		{object}	httperr.ErrorResponse
// @Failure			401		{object}	httperr.ErrorResponse
// @Failure			404		{object}	httperr.ErrorResponse
// @Failure			500		{object}	httperr.ErrorResponse
// @Router			/orders/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httperr.WriteError(c, http.StatusBadRequest, "invalid order id")
		return
	}

	in, ok := ginbind.JSON[UpdateOrderDTO](c, "invalid json body")
	if !ok {
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
		httperr.WriteError(c, http.StatusNotFound, "order not found")
	case errors.Is(err, ErrServiceNotInOrder):
		httperr.WriteError(c, http.StatusBadRequest, "order service does not belong to order")
	default:
		httperr.WriteValidationOrInternal(c, err)
	}
}
