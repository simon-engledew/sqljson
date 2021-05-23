package types

type CreateColumn struct {
	Type string `json:"type"`
}

type CreateTable struct {
	Columns map[string]*CreateColumn `json:"columns"`
}
