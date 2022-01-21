package main

import (
	"embed"
	"encoding/json"
	"flag"
	"github.com/jpillora/longestcommon"
	"github.com/simon-engledew/sqljson/internal/data"
	"github.com/simon-engledew/sqljson/internal/relationships"
	"hash/fnv"
	"io"
	"os"
	"text/template"
)

//go:embed templates
var content embed.FS

var colors = map[int][]string{
	100: {
		"#FFCDD2",
		"#F8BBD0",
		"#E1BEE7",
		"#D1C4E9",
		"#C5CAE9",
		"#BBDEFB",
		"#B3E5FC",
		"#B2EBF2",
		"#B2DFDB",
		"#C8E6C9",
		"#DCEDC8",
		"#F0F4C3",
		"#FFF9C4",
		"#FFECB3",
		"#FFE0B2",
		"#FFCCBC",
		"#D7CCC8",
		"#F5F5F5",
		"#CFD8DC",
	},
	300: {
		"#E57373",
		"#F06292",
		"#BA68C8",
		"#9575CD",
		"#7986CB",
		"#64B5F6",
		"#4FC3F7",
		"#4DD0E1",
		"#4DB6AC",
		"#81C784",
		"#AED581",
		"#DCE775",
		"#FFF176",
		"#FFD54F",
		"#FFB74D",
		"#FF8A65",
		"#A1887F",
		"#E0E0E0",
		"#90A4AE",
	},
}

func Hash(key string) (int, error) {
	h := fnv.New32a()
	if _, err := h.Write([]byte(key)); err != nil {
		return 0, err
	}
	return int(h.Sum32()), nil
}

func Color(level int, index int) string {
	items := colors[level]
	return items[index%len(items)]
}

func Transform(r io.Reader, w io.Writer, format string) error {
	templates := map[string]*template.Template{
		"dot": template.Must(template.New("dot").Funcs(template.FuncMap{
			"Color":  Color,
			"Hash":   Hash,
			"Escape": template.HTMLEscapeString,
		}).ParseFS(content, "templates/dot/*")),
		"mermaidjs": template.Must(template.New("mermaidjs").Funcs(template.FuncMap{
			"Color":  Color,
			"Hash":   Hash,
			"Escape": template.HTMLEscapeString,
		}).ParseFS(content, "templates/mermaidjs/*")),
	}

	var createTables map[string]*data.CreateTable

	dec := json.NewDecoder(r)
	if err := dec.Decode(&createTables); err != nil {
		return err
	}

	tableNames := make([]string, 0, len(createTables))
	for tableName := range createTables {
		tableNames = append(tableNames, tableName)
	}

	prefix := longestcommon.Prefix(tableNames)

	relatedTables := relationships.Find(createTables, relationships.WithPrefix(prefix, relationships.ForeignKey))

	return templates[format].ExecuteTemplate(w, "main.tmpl", &relatedTables)
}

func main() {
	format := flag.String("format", "dot", "diagram format")
	flag.Parse()

	err := Transform(os.Stdin, os.Stdout, *format)
	if err != nil {
		panic(err)
	}
}
