// Package schema is used to provide necessary functionality for representing
// a JSON schema as a nested struct
package schema

import "encoding/json"

// DataType represents a string encoding of different datatypes
type DataType string

const (
	Object  DataType = "object"
	Number  DataType = "number"
	Integer DataType = "integer"
	String  DataType = "string"
	Array   DataType = "array"
	Null    DataType = "null"
	Boolean DataType = "boolean"
)

// Definition is a struct for describing a JSON Schema
type Definition struct {
	Type        DataType              `json:"type,omitempty"`        // Type specifies the data type of the schema
	Description string                `json:"description,omitempty"` // Description is the description of the schema
	Enum        []string              `json:"enum,omitempty"`        // Enum is used to restrict a value to a fixed set of values (must have at least one element with unique values)
	Properties  map[string]Definition `json:"properties"`            // Properties describes the properties of an object, if the schema type is an object
	Required    []string              `json:"required,omitempty"`    // Required specifies which properties are required, if the schema type is an object
	Items       *Definition           `json:"items,omitempty"`       // Items specifies which data type an array contains, if the schema type is an array
}

// Marshal a struct into JSON
func (d Definition) MarshalJSON() ([]byte, error) {
	if d.Properties == nil {
		// Make an empty map for the definition if it has no properties
		d.Properties = make(map[string]Definition)
	}

	// Return a marshaled struct
	type Alias Definition
	return json.Marshal(struct{ Alias }{Alias: (Alias)(d)})
}
