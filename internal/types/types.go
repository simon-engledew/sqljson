package types

type CreateColumn struct {
	Type string `json:"type"`
}

type CreateTable struct {
	Columns       map[string]*CreateColumn `json:"columns"`
	Relationships map[string]string        `json:"relationships"`
}
