package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap/parser"
	_ "github.com/pingcap/parser/test_driver"
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

func main() {
	db, err := sql.Open("mysql", "root:@tcp(localhost:13306)/information_schema")
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	var schemas Schemas

	must(db.QueryRow("SELECT JSON_ARRAYAGG(JSON_OBJECT('catalog', sys.quote_identifier(CATALOG_NAME), 'name', sys.quote_identifier(SCHEMA_NAME))) FROM information_schema.schemata").Scan(&schemas))

	p := parser.New()

	for _, schema := range schemas {
		var name, statement string
		must(db.QueryRow(fmt.Sprintf(`SHOW CREATE SCHEMA %s`, schema.Name)).Scan(&name, &statement))
		fmt.Println(statement)

		var tables Tables
		must(db.QueryRow("SELECT JSON_ARRAYAGG(JSON_OBJECT('schema', sys.quote_identifier(TABLE_SCHEMA), 'name', sys.quote_identifier(TABLE_NAME))) FROM information_schema.tables WHERE table_type<>'VIEW' AND table_schema = ?", name).Scan(&tables))

		for _, table := range tables {
			must(db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s.%s", table.Schema, table.Name)).Scan(&name, &statement))
			fmt.Println(statement)

			stmtNodes, _, err := p.Parse(statement, "", "")
			if err != nil {
				panic(err)
			}
			fmt.Println(stmtNodes[0])
		}
	}
}
