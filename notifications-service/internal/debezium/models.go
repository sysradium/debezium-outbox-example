package debezium

import "encoding/json"

type Root struct {
	Schema  Schema  `json:"schema"`
	Payload Payload `json:"payload"`
}

type Schema struct {
	Type     string  `json:"type"`
	Fields   []Field `json:"fields"`
	Optional bool    `json:"optional"`
	Name     string  `json:"name"`
}

type Field struct {
	Type     string  `json:"type"`
	Optional bool    `json:"optional"`
	Field    string  `json:"field"`
	Name     string  `json:"name,omitempty"`
	Version  int     `json:"version,omitempty"`
	Fields   []Field `json:"fields,omitempty"`
}

type Payload struct {
	Payload json.RawMessage `json:"payload"`
	Id      string          `json:"id"`
}
