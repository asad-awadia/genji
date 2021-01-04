package stream_test

import (
	"fmt"
	"testing"

	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/sql/parser"
	"github.com/genjidb/genji/sql/query/expr"
	"github.com/genjidb/genji/stream"
	"github.com/stretchr/testify/require"
)

func TestStream(t *testing.T) {
	s := stream.New(stream.NewDocumentIterator(
		docFromJSON(`{"a": 1}`),
		docFromJSON(`{"a": 2}`),
	))

	s = s.Pipe(stream.Map(parser.MustParseExpr("{a: a + 1}")))
	s = s.Pipe(stream.Filter(parser.MustParseExpr("a > 2")))

	var count int64
	err := s.Iterate(func(env *expr.Environment) error {
		d, ok := env.GetDocument()
		require.True(t, ok)
		require.JSONEq(t, fmt.Sprintf(`{"a": %d}`, count+3), document.NewDocumentValue(d).String())
		count++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}