package importer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/pingcap/tidb/ast"
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/util/types"
)

type Table struct {
	name         string
	columns      []*column
	columnList   string
	indices      map[string]*column
	uniqIndices  map[string]*column
	unsignedCols map[string]*column
}

func (t *Table) parseTableConstraint(cons *ast.Constraint) {
	switch cons.Tp {
	case ast.ConstraintPrimaryKey, ast.ConstraintKey, ast.ConstraintUniq,
		ast.ConstraintUniqKey, ast.ConstraintUniqIndex:
		for _, indexCol := range cons.Keys {
			name := indexCol.Column.Name.L
			t.uniqIndices[name] = t.findCol(t.columns, name)
		}
	case ast.ConstraintIndex:
		for _, indexCol := range cons.Keys {
			name := indexCol.Column.Name.L
			t.indices[name] = t.findCol(t.columns, name)
		}
	}
}

func (t *Table) findCol(cols []*column, name string) *column {
	for _, col := range cols {
		if col.name == name {
			return col
		}
	}

	return nil
}

func (t *Table) buildColumnList() {
	columns := make([]string, 0, len(t.columns))
	for _, column := range t.columns {
		columns = append(columns, column.name)
	}

	t.columnList = strings.Join(columns, ",")
}

type column struct {
	idx     int
	name    string
	data    *datum
	tp      *types.FieldType
	comment string
	min     string
	max     string
	step    int64
	set     []string

	table *Table
}

func (col *column) String() string {
	if col == nil {
		return "<nil>"
	}

	return fmt.Sprintf("[column]idx: %d, name: %s, tp: %v, min: %s, max: %s, step: %d, set: %v\n",
		col.idx, col.name, col.tp, col.min, col.max, col.step, col.set)
}

func (col *column) parseColumn(cd *ast.ColumnDef) {
	col.name = cd.Name.Name.L
	col.tp = cd.Tp
	col.parseColumnOptions(cd.Options)
	col.parseColumnComment()
	col.table.columns = append(col.table.columns, col)
}

func (col *column) parseColumnOptions(ops []*ast.ColumnOption) {
	for _, op := range ops {
		switch op.Tp {
		case ast.ColumnOptionPrimaryKey, ast.ColumnOptionUniq, ast.ColumnOptionAutoIncrement,
			ast.ColumnOptionUniqIndex, ast.ColumnOptionKey, ast.ColumnOptionUniqKey:
			col.table.uniqIndices[col.name] = col
		case ast.ColumnOptionIndex:
			col.table.indices[col.name] = col
		case ast.ColumnOptionComment:
			col.comment = op.Expr.GetDatum().GetString()
		}
	}
}

// parse the data rules.
// rule like `a int unique comment '[[range=1,10;step=1]]'`,
// then we will get value from 1,2...10
func (col *column) parseColumnComment() {
	comment := strings.TrimSpace(col.comment)
	start := strings.Index(comment, "[[")
	end := strings.Index(comment, "]]")
	var content string
	if start < end {
		content = comment[start+2 : end]
	}

	fields := strings.Split(content, ";")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		kvs := strings.Split(field, "=")
		col.parseRule(kvs)
	}
}

func (col *column) parseRule(kvs []string) {
	if len(kvs) != 2 {
		return
	}

	key := strings.TrimSpace(kvs[0])
	value := strings.TrimSpace(kvs[1])
	if key == "range" {
		fields := strings.Split(value, ",")
		if len(fields) == 1 {
			col.min = strings.TrimSpace(fields[0])
		} else if len(fields) == 2 {
			col.min = strings.TrimSpace(fields[0])
			col.max = strings.TrimSpace(fields[1])
		}
	} else if key == "step" {
		var err error
		col.step, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
	} else if key == "set" {
		fields := strings.Split(value, ",")
		for _, field := range fields {
			col.set = append(col.set, strings.TrimSpace(field))
		}
	}
}

// create table
func NewTable() *Table {
	return &Table{
		indices:      make(map[string]*column),
		uniqIndices:  make(map[string]*column),
		unsignedCols: make(map[string]*column),
	}
}

// Parse create table SQL
func ParseTableSQL(table *Table, sql string) error {
	stmt, err := parser.New().ParseOneStmt(sql, "", "")
	if err != nil {
		return errors.Trace(err)
	}

	switch node := stmt.(type) {
	case *ast.CreateTableStmt:
		err = parseTable(table, node)
	default:
		err = errors.Errorf("invalid statement - %v", stmt.Text())
	}

	return errors.Trace(err)
}

func parseTable(t *Table, stmt *ast.CreateTableStmt) error {
	t.name = stmt.Table.Name.L
	t.columns = make([]*column, 0, len(stmt.Cols))

	for i, col := range stmt.Cols {
		column := &column{idx: i + 1, table: t, step: defaultStep, data: newDatum()}
		column.parseColumn(col)
	}

	for _, cons := range stmt.Constraints {
		t.parseTableConstraint(cons)
	}

	t.buildColumnList()

	return nil
}

// Parse create index SQL
func ParseIndexSQL(table *Table, sql string) error {
	if len(sql) == 0 {
		return nil
	}

	stmt, err := parser.New().ParseOneStmt(sql, "", "")
	if err != nil {
		return errors.Trace(err)
	}

	switch node := stmt.(type) {
	case *ast.CreateIndexStmt:
		err = parseIndex(table, node)
	default:
		err = errors.Errorf("invalid statement - %v", stmt.Text())
	}

	return errors.Trace(err)
}

func parseIndex(table *Table, stmt *ast.CreateIndexStmt) error {
	if table.name != stmt.Table.Name.L {
		return errors.Errorf("mismatch table name for create index - %s : %s", table.name, stmt.Table.Name.L)
	}

	for _, indexCol := range stmt.IndexColNames {
		name := indexCol.Column.Name.L
		if stmt.Unique {
			table.uniqIndices[name] = table.findCol(table.columns, name)
		} else {
			table.indices[name] = table.findCol(table.columns, name)
		}
	}

	return nil
}
