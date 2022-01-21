package data

type CreateColumn struct {
	Type string `json:"type"`
	Kind string `json:"kind"`
}

type CreateTable struct {
	Columns map[string]*CreateColumn `json:"columns"`
}
