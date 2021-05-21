package main

import (
	"embed"
	"fmt"
	"hash/fnv"
	"html"
	"text/template"
)

//go:embed templates/*
var content embed.FS

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

func Column(column CreateColumn) string {
	if column.Relationship {
		return fmt.Sprintf("<U>%s</U>", html.EscapeString(column.Name))
	}
	return html.EscapeString(column.Name)
}

func Color(key string) string {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := int(h.Sum32()) % len(colors)

	return colors[idx]
}

var dot = func() *template.Template {
	return template.Must(template.New("dot").Funcs(template.FuncMap{"Color": Color, "Column": Column}).ParseFS(content, "templates/*"))
}()
