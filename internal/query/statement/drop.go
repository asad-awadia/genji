package statement

import (
	"fmt"

	errs "github.com/chaisql/chai/internal/errors"
	"github.com/cockroachdb/errors"
)

// DropTableStmt is a DSL that allows creating a DROP TABLE query.
type DropTableStmt struct {
	TableName string
	IfExists  bool
}

// IsReadOnly always returns false. It implements the Statement interface.
func (stmt DropTableStmt) IsReadOnly() bool {
	return false
}

// Run runs the DropTable statement in the given transaction.
// It implements the Statement interface.
func (stmt DropTableStmt) Run(ctx *Context) (Result, error) {
	var res Result

	if stmt.TableName == "" {
		return res, errors.New("missing table name")
	}

	tb, err := ctx.Tx.Catalog.GetTable(ctx.Tx, stmt.TableName)
	if err != nil {
		if errs.IsNotFoundError(err) && stmt.IfExists {
			err = nil
		}

		return res, err
	}

	err = ctx.Tx.CatalogWriter().DropTable(ctx.Tx, stmt.TableName)
	if err != nil {
		return res, err
	}

	// if there is no primary key, drop the rowid sequence
	if tb.Info.PrimaryKey == nil {
		err = ctx.Tx.CatalogWriter().DropSequence(ctx.Tx, tb.Info.RowidSequenceName)
		if err != nil {
			return res, err
		}
	}

	return res, err
}

// DropIndexStmt is a DSL that allows creating a DROP INDEX query.
type DropIndexStmt struct {
	IndexName string
	IfExists  bool
}

// IsReadOnly always returns false. It implements the Statement interface.
func (stmt DropIndexStmt) IsReadOnly() bool {
	return false
}

// Run runs the DropIndex statement in the given transaction.
// It implements the Statement interface.
func (stmt DropIndexStmt) Run(ctx *Context) (Result, error) {
	var res Result

	if stmt.IndexName == "" {
		return res, errors.New("missing index name")
	}

	err := ctx.Tx.CatalogWriter().DropIndex(ctx.Tx, stmt.IndexName)
	if errs.IsNotFoundError(err) && stmt.IfExists {
		err = nil
	}

	return res, err
}

// DropSequenceStmt is a DSL that allows creating a DROP INDEX query.
type DropSequenceStmt struct {
	SequenceName string
	IfExists     bool
}

// IsReadOnly always returns false. It implements the Statement interface.
func (stmt DropSequenceStmt) IsReadOnly() bool {
	return false
}

// Run runs the DropSequence statement in the given transaction.
// It implements the Statement interface.
func (stmt DropSequenceStmt) Run(ctx *Context) (Result, error) {
	var res Result

	if stmt.SequenceName == "" {
		return res, errors.New("missing index name")
	}

	seq, err := ctx.Tx.Catalog.GetSequence(stmt.SequenceName)
	if err != nil {
		if errs.IsNotFoundError(err) && stmt.IfExists {
			err = nil
		}
		return res, err
	}

	if seq.Info.Owner.TableName != "" {
		return res, fmt.Errorf("cannot drop sequence %s because constraint of table %s requires it", seq.Info.Name, seq.Info.Owner.TableName)
	}

	err = ctx.Tx.CatalogWriter().DropSequence(ctx.Tx, stmt.SequenceName)
	if errs.IsNotFoundError(err) && stmt.IfExists {
		err = nil
	}

	return res, err
}
