package pagination

import "math"

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Query representa os parâmetros comuns de paginação.
type Query struct {
	Page     int `form:"page" json:"page" validate:"min=1"`
	PageSize int `form:"page_size" json:"page_size" validate:"min=1,max=100"`
}

// Meta representa os metadados comuns de uma resposta paginada.
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Normalize aplica defaults e limites máximos aos parâmetros.
func Normalize(q Query) Query {
	if q.Page == 0 {
		q.Page = DefaultPage
	}
	if q.PageSize == 0 {
		q.PageSize = DefaultPageSize
	}
	if q.PageSize > MaxPageSize {
		q.PageSize = MaxPageSize
	}
	return q
}

// Offset calcula o deslocamento para queries SQL.
func Offset(q Query) int {
	return (q.Page - 1) * q.PageSize
}

// TotalPages calcula a quantidade total de páginas para uma página de tamanho fixo.
func TotalPages(total int64, pageSize int) int {
	if total == 0 || pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}

// NewMeta monta os metadados de uma resposta paginada.
func NewMeta(q Query, total int64) Meta {
	return Meta{
		Page:       q.Page,
		PageSize:   q.PageSize,
		Total:      total,
		TotalPages: TotalPages(total, q.PageSize),
	}
}
