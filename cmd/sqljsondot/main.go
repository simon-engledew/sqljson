package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/simon-engledew/sqlinspect/internal/types"
	"hash/fnv"
	"html"
	"io"
	"os"
	"text/template"
)

//go:embed templates/*
var content embed.FS

var colors = []string{
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
}

func Underline(text string) string {
	return fmt.Sprintf("<U>%s</U>", html.EscapeString(text))
}

func Color(key string) (string, error) {
	h := fnv.New32a()
	if _, err := h.Write([]byte(key)); err != nil {
		return "", err
	}
	return ColorAt(int(h.Sum32())), nil
}

func ColorAt(index int) string {
	return colors[index%len(colors)]
}

func Convert(r io.Reader, w io.Writer) error {
	dot, err := template.New("dot").Funcs(template.FuncMap{"Color": Color, "ColorAt": ColorAt, "Underline": Underline}).ParseFS(content, "templates/*")
	if err != nil {
		return err
	}

	var createSchema types.CreateSchema

	dec := json.NewDecoder(r)
	if err := dec.Decode(&createSchema); err != nil {
		return err
	}

	return dot.ExecuteTemplate(w, "dot.tmpl", &createSchema)
}

func main() {
	err := Convert(os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}
}