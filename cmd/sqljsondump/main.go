package main

import (
	"encoding/json"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/simon-engledew/sqljson/internal/types"
	"io"

	_ "github.com/pingcap/parser/test_driver"

	"io/ioutil"
	"os"
)

// Transform returns a function that will read a MySQL dump from r and write a JSON description to w.
func Transform() func(r io.Reader, w io.Writer) error {
	p := parser.New()

	return func(r io.Reader, w io.Writer) error {
		dump, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		statements, _, err := p.Parse(string(dump), "", "")
		if err != nil {
			return err
		}

		tables := make(map[string]*types.CreateTable)

		for _, statement := range statements {
			if create, ok := statement.(*ast.CreateTableStmt); ok {
				tableName := create.Table.Name.String()

				createTable := &types.CreateTable{
					Columns: make(map[string]*types.CreateColumn),
				}

				for _, col := range create.Cols {
					columnName := col.Name.String()

					createColumn := &types.CreateColumn{
						Type: col.Tp.InfoSchemaStr(),
					}
					createTable.Columns[columnName] = createColumn
				}

				tables[tableName] = createTable
			}
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		return enc.Encode(&tables)
	}
}

func main() {
	if err := Transform()(os.Stdin, os.Stdout); err != nil {
		panic(err)
	}
}
