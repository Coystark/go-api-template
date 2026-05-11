package pagination

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize_AppliesDefaultsAndCapsPageSize(t *testing.T) {
	out := Normalize(Query{PageSize: 500})

	require.Equal(t, DefaultPage, out.Page)
	require.Equal(t, MaxPageSize, out.PageSize)
}

func TestOffset(t *testing.T) {
	require.Equal(t, 40, Offset(Query{Page: 3, PageSize: 20}))
}

func TestNewMeta(t *testing.T) {
	out := NewMeta(Query{Page: 2, PageSize: 20}, 41)

	require.Equal(t, 2, out.Page)
	require.Equal(t, 20, out.PageSize)
	require.Equal(t, int64(41), out.Total)
	require.Equal(t, 3, out.TotalPages)
}
