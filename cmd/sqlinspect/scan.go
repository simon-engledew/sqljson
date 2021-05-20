package main

import (
	"database/sql/driver"
	"encoding/json"
)

type Schemas []*Schema

func (s *Schemas) Scan(val interface{}) error {
	data := val.([]uint8)
	return json.Unmarshal(data, &s)
}

func (s *Schemas) Value() (driver.Value, error) {
	return json.Marshal(&s)
}

type Tables []*Table

func (s *Tables) Scan(val interface{}) error {
	data := val.([]uint8)
	return json.Unmarshal(data, &s)
}

func (s *Tables) Value() (driver.Value, error) {
	return json.Marshal(&s)
}
