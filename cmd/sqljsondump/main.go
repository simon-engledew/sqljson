package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/inflection"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/parser/test_driver"
	"github.com/simon-engledew/sqlinspect/internal/types"
	"os"
	"strings"
	"time"
)

type Schema struct {
	Catalog string `json:"catalog"`
	Name    string `json:"name"`
}

type Table struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

func (s *Schema) String() string {
	return s.Catalog + `.` + s.Name
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func quoteIdentifier(db *sql.DB, identifier string) (quoted string) {
	must(db.QueryRow("SELECT sys.quote_identifier(?)", identifier).Scan(&quoted))
	return
}

func main() {
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "",
		Net:                  "tcp",
		Addr:                 ":13306",
		DBName:               "information_schema",
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	var schemas Schemas

	must(db.QueryRow(`
SELECT
    JSON_ARRAYAGG(
        JSON_OBJECT(
            'catalog', CATALOG_NAME,
            'name', SCHEMA_NAME
        )
    )
FROM information_schema.schemata`).Scan(&schemas))

	p := parser.New()

	prefix := os.Args[2]

	for _, schema := range schemas {
		// TODO
		if schema.Name != os.Args[1] {
			continue
		}

		var name, statement string
		must(db.QueryRow(fmt.Sprintf(`SHOW CREATE SCHEMA %s`, quoteIdentifier(db, schema.Name))).Scan(&name, &statement))

		var tables Tables
		must(db.QueryRow(`
SELECT
	JSON_ARRAYAGG(
		JSON_OBJECT(
			'schema', TABLE_SCHEMA,
			'name', TABLE_NAME
		)
	)
FROM information_schema.tables
WHERE table_type <> 'VIEW' AND table_schema = ?`, name).Scan(&tables))

		createSchema := types.CreateSchema{
			Name:   schema.Name,
			Tables: make([]*types.CreateTable, 0, len(tables)),
		}

		tableNames := make(map[string]struct{})

		for _, table := range tables {
			must(db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s.%s", quoteIdentifier(db, table.Schema), quoteIdentifier(db, table.Name))).Scan(&name, &statement))

			nodes, _, err := p.Parse(statement, "", "")
			if err != nil {
				panic(err)
			}
			for _, node := range nodes {
				create := node.(*ast.CreateTableStmt)

				createTable := types.CreateTable{
					Name:          create.Table.Name.String(),
					Columns:       make([]*types.CreateColumn, 0, len(create.Cols)),
					Relationships: make(map[string]string, 0),
				}

				tableNames[createTable.Name] = struct{}{}

				for _, col := range create.Cols {
					createColumn := &types.CreateColumn{
						Name: col.Name.String(),
						Type: col.Tp.InfoSchemaStr(),
					}
					createTable.Columns = append(createTable.Columns, createColumn)
				}

				createSchema.Tables = append(createSchema.Tables, &createTable)
			}
		}

		for _, table := range createSchema.Tables {
			for _, column := range table.Columns {
				if strings.HasSuffix(column.Name, "_id") {
					target := prefix + inflection.Plural(column.Name[:len(column.Name)-3])

					if _, ok := tableNames[target]; ok {
						table.Relationships[column.Name] = target
					}
				}
			}
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", " ")
		must(enc.Encode(&createSchema))
	}
}
