package types

type CreateColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type CreateTable struct {
	Name          string            `json:"name"`
	Columns       []*CreateColumn   `json:"columns"`
	Relationships map[string]string `json:"relationships"`
}

type CreateSchema struct {
	Name   string         `json:"name"`
	Tables []*CreateTable `json:"tables"`
}
