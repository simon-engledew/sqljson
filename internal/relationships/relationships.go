package relationships

import (
	"github.com/jinzhu/inflection"
	"github.com/simon-engledew/sqljson/internal/types"
	"strings"
)

// RelatedTo will return a tableName if the column has some relationship to it
type RelatedTo func(columnName string) (tableName string, ok bool)

// WithPrefix adds a prefix to any tables returned by relatedTo
func WithPrefix(prefix string, relatedTo RelatedTo) RelatedTo {
	return func(columnName string) (tableName string, ok bool) {
		tableName, ok = relatedTo(columnName)
		if ok {
			return prefix + tableName, ok
		}
		return
	}
}

// ForeignKey matches columns of the format <singular tableName>_id
func ForeignKey(columnName string) (tableName string, ok bool) {
	if strings.HasSuffix(string(columnName), "_id") {
		return inflection.Plural(columnName[:len(columnName)-3]), true
	}
	return "", false
}

// CreateTable with an additional map containing relationships to other tables.
type CreateTable struct {
	types.CreateTable
	Relationships map[string]string
}

// Find returns an extended version of tables which contains any relationships described by the function relatedTo
func Find(tables map[string]*types.CreateTable, relatedTo RelatedTo) map[string]*CreateTable {
	relationships := make(map[string]*CreateTable)
	for tableName, table := range tables {
		current := &CreateTable{
			CreateTable:   *table,
			Relationships: make(map[string]string),
		}
		relationships[tableName] = current
		for columnName, _ := range table.Columns {
			if target, ok := relatedTo(columnName); ok {
				if _, ok := tables[target]; ok {
					current.Relationships[columnName] = target
				}
			}
		}
	}
	return relationships
}
