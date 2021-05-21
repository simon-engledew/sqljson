package main

import (
	"encoding/json"
	"flag"
	"github.com/jinzhu/inflection"
	"strings"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/simon-engledew/sqlinspect/internal/types"

	_ "github.com/pingcap/parser/test_driver"

	"io/ioutil"
	"os"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var prefixFlag = flag.String("prefix", "", "Table prefix")

func main() {
	flag.Parse()

	p := parser.New()

	prefix := *prefixFlag

	dump, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	statements, _, err := p.Parse(string(dump), "", "")
	if err != nil {
		panic(err)
	}

	tables := make(map[string]*types.CreateTable)

	for _, statement := range statements {
		if create, ok := statement.(*ast.CreateTableStmt); ok {
			tableName := create.Table.Name.String()

			createTable := &types.CreateTable{
				Columns:       make(map[string]*types.CreateColumn),
				Relationships: make(map[string]string),
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

	for _, table := range tables {
		for columnName, _ := range table.Columns {
			if strings.HasSuffix(columnName, "_id") {
				target := prefix + inflection.Plural(columnName[:len(columnName)-3])

				if _, ok := tables[target]; ok {
					table.Relationships[columnName] = target
				}
			}
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	must(enc.Encode(&tables))
}
