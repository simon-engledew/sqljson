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
	"io/ioutil"
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

type CreateColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type CreateTable struct {
	Name    string         `json:"name"`
	Columns []CreateColumn `json:"columns"`
}

type CreateSchema struct {
	Name   string        `json:"name"`
	Tables []CreateTable `json:"tables"`
}

func quoteIdentifier(db *sql.DB, identifier string) (quoted string) {
	must(db.QueryRow("SELECT sys.quote_identifier(?)", identifier).Scan(&quoted))
	return
}

var colors = []string{
	"#FFEBEE",
	"#FCE4EC",
	"#F3E5F5",
	"#EDE7F6",
	"#E8EAF6",
	"#E3F2FD",
	"#E1F5FE",
	"#E0F7FA",
	"#E0F2F1",
	"#E8F5E9",
	"#F1F8E9",
	"#F9FBE7",
	"#FFFDE7",
	"#FFF8E1",
	"#FFF3E0",
	"#FBE9E7",
	"#EFEBE9",
	"#FAFAFA",
	"#ECEFF1",
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
		// fmt.Println(statement)

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

		createSchema := CreateSchema{
			Name:   schema.Name,
			Tables: make([]CreateTable, 0, len(tables)),
		}

		tableNames := make(map[string]struct{})
		relationships := make(map[string][]string)

		for i, table := range tables {
			must(db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s.%s", quoteIdentifier(db, table.Schema), quoteIdentifier(db, table.Name))).Scan(&name, &statement))
			// fmt.Println(statement)

			nodes, _, err := p.Parse(statement, "", "")
			if err != nil {
				panic(err)
			}
			for _, node := range nodes {
				create := node.(*ast.CreateTableStmt)

				createTable := CreateTable{
					Name:    create.Table.Name.String(),
					Columns: make([]CreateColumn, 0, len(create.Cols)),
				}

				tableNames[createTable.Name] = struct{}{}

				fmt.Printf("[%s] {bgcolor: %q}\n", createTable.Name, colors[i%len(colors)])

				for _, col := range create.Cols {
					createColumn := CreateColumn{
						Name: col.Name.String(),
						Type: col.Tp.InfoSchemaStr(),
					}
					createTable.Columns = append(createTable.Columns, createColumn)

					if strings.HasSuffix(createColumn.Name, "_id") && strings.EqualFold(createColumn.Type, "bigint(20) unsigned") {
						relationships[createTable.Name] = append(relationships[createTable.Name], prefix+inflection.Plural(createColumn.Name[:len(createColumn.Name)-3]))
					}

					fmt.Printf("  %s {label: %q}\n", createColumn.Name, createColumn.Type)
					//fmt.Println(col.Name, col.Tp.InfoSchemaStr())
					//for _, opt := range col.Options {
					//	switch opt.Tp {
					//	//case ast.ColumnOptionUniqKey:
					//	case ast.ColumnOptionNotNull:
					//		fmt.Println("NN")
					//	case ast.ColumnOptionAutoIncrement:
					//		fmt.Println("AI")
					//	case ast.ColumnOptionPrimaryKey:
					//		fmt.Println("*")
					//	}
					//}

					//for _, option := range col.Options {
					//	fmt.Println(option.Expr)
					//}
				}

				createSchema.Tables = append(createSchema.Tables, createTable)
			}

			fmt.Println()
		}

		for source, targets := range relationships {
			if _, ok := tableNames[source]; ok {
				for _, target := range targets {
					if _, ok := tableNames[target]; ok {
						fmt.Println(source, "1--*", target)
					}
				}
			}
		}

		enc := json.NewEncoder(ioutil.Discard)
		enc.SetIndent("", " ")
		enc.Encode(&createSchema)
	}
}
