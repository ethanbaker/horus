package schema_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ethanbaker/horus/utils/schema"
	"github.com/stretchr/testify/assert"
)

// Path to expected output file
const EXPECTED_PATH = "./testing/expected.json"

// test type holds a given test and test name to find and match expected output
type test struct {
	name string            // The name of the test
	def  schema.Definition // The definition being marshaled in the test
}

// Compile a list of all tests and initialize the name and definition
var TESTS = []test{
	{
		name: "Test with empty definition",
		def:  schema.Definition{},
	},
	{
		name: "Test with definition properties set",
		def: schema.Definition{
			Type:        schema.String,
			Description: "A string type",
			Properties: map[string]schema.Definition{
				"name": {
					Type: schema.String,
				},
			},
		},
	},
	{
		name: "Test with nested definition properties",
		def: schema.Definition{
			Type: schema.Object,
			Properties: map[string]schema.Definition{
				"name": {
					Type: schema.String,
				},
				"age": {
					Type: schema.Integer,
				},
			},
		},
	},
	{
		name: "Test with complex nested definition",
		def: schema.Definition{
			Type: schema.Object,
			Properties: map[string]schema.Definition{
				"user": {
					Type: schema.Object,
					Properties: map[string]schema.Definition{
						"name": {
							Type: schema.String,
						},
						"age": {
							Type: schema.Integer,
						},
						"address": {
							Type: schema.Object,
							Properties: map[string]schema.Definition{
								"city": {
									Type: schema.String,
								},
								"country": {
									Type: schema.String,
								},
							},
						},
					},
				},
			},
		},
	},
	{
		name: "Test with array type definition",
		def: schema.Definition{
			Type: schema.Array,
			Items: &schema.Definition{
				Type: schema.String,
			},
			Properties: map[string]schema.Definition{
				"name": {
					Type: schema.String,
				},
			},
		},
	},
}

func TestMarshalJSON(t *testing.T) {
	assert := assert.New(t)

	// Open the testing file with expected values
	file, err := os.ReadFile(EXPECTED_PATH)
	assert.Nil(err)

	// Unmarshal the expected values into a map
	var expected map[string]json.RawMessage
	assert.Nil(json.Unmarshal(file, &expected))

	// Get rid of stored JSON artifacts (backslashes and leading/ending quotes)
	for key := range expected {
		expected[key] = json.RawMessage(strings.ReplaceAll(string(expected[key]), "\\", ""))
		expected[key] = json.RawMessage(strings.Trim(string(expected[key]), `"`))
	}

	// Loop through each test and expect the output
	for _, test := range TESTS {
		want := expected[test.name]

		got, err := test.def.MarshalJSON()
		assert.Nil(err)

		assert.Equal([]byte(want), got, test.name)
	}
}
