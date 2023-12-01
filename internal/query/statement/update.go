package statement

import (
	"github.com/cockroachdb/errors"
	"github.com/genjidb/genji/internal/expr"
	"github.com/genjidb/genji/internal/object"
	"github.com/genjidb/genji/internal/stream"
	"github.com/genjidb/genji/internal/stream/index"
	"github.com/genjidb/genji/internal/stream/path"
	"github.com/genjidb/genji/internal/stream/rows"
	"github.com/genjidb/genji/internal/stream/table"
)

// UpdateConfig holds UPDATE configuration.
type UpdateStmt struct {
	basePreparedStatement

	TableName string

	// SetPairs is used along with the Set clause. It holds
	// each path with its corresponding value that
	// should be set in the object.
	SetPairs []UpdateSetPair

	// UnsetFields is used along with the Unset clause. It holds
	// each path that should be unset from the object.
	UnsetFields []string

	WhereExpr expr.Expr
}

func NewUpdateStatement() *UpdateStmt {
	var p UpdateStmt

	p.basePreparedStatement = basePreparedStatement{
		Preparer: &p,
		ReadOnly: false,
	}

	return &p
}

type UpdateSetPair struct {
	Path object.Path
	E    expr.Expr
}

// Prepare implements the Preparer interface.
func (stmt *UpdateStmt) Prepare(c *Context) (Statement, error) {
	ti, err := c.Tx.Catalog.GetTableInfo(stmt.TableName)
	if err != nil {
		return nil, err
	}
	pk := ti.GetPrimaryKey()

	s := stream.New(table.Scan(stmt.TableName))

	if stmt.WhereExpr != nil {
		s = s.Pipe(rows.Filter(stmt.WhereExpr))
	}

	var pkModified bool
	if stmt.SetPairs != nil {
		for _, pair := range stmt.SetPairs {
			// if we modify the primary key,
			// we must remove the old row and create an new one
			if pk != nil && !pkModified {
				for _, p := range pk.Paths {
					if p.IsEqual(pair.Path) {
						pkModified = true
						break
					}
				}
			}
			s = s.Pipe(path.Set(pair.Path, pair.E))
		}
	} else if stmt.UnsetFields != nil {
		for _, name := range stmt.UnsetFields {
			// ensure we do not unset any path the is used in the primary key
			if pk != nil {
				path := object.NewPath(name)
				for _, p := range pk.Paths {
					if p.IsEqual(path) {
						return nil, errors.New("cannot unset primary key path")
					}
				}
			}
			s = s.Pipe(path.Unset(name))
		}
	}

	// validate row
	s = s.Pipe(table.Validate(stmt.TableName))

	// TODO(asdine): This removes ALL indexed fields for each row
	// even if the update modified a single field. We should only
	// update the indexed fields that were modified.
	indexNames := c.Tx.Catalog.ListIndexes(stmt.TableName)
	for _, indexName := range indexNames {
		s = s.Pipe(index.Delete(indexName))
	}

	if pkModified {
		s = s.Pipe(table.Delete(stmt.TableName))
		s = s.Pipe(table.Insert(stmt.TableName))
	} else {
		s = s.Pipe(table.Replace(stmt.TableName))
	}

	for _, indexName := range indexNames {
		info, err := c.Tx.Catalog.GetIndexInfo(indexName)
		if err != nil {
			return nil, err
		}
		if info.Unique {
			s = s.Pipe(index.Validate(indexName))
		}

		s = s.Pipe(index.Insert(indexName))
	}

	s = s.Pipe(stream.Discard())

	st := StreamStmt{
		Stream:   s,
		ReadOnly: false,
	}

	return st.Prepare(c)
}
