package graphql

import (
	"fmt"
	"jgschema/jsonutils"
	"path/filepath"
	"reflect"
	"testing"
)

// TestTransform tests the private 'transform' function.
func TestTransform(t *testing.T) {
	type test struct {
		description string
		inputSchema string // path to test file
		wantGraphQL []Schema
		wantErr     error
	}

	schemaTestDir := "./test_data/jsonschema"
	tests := []test{
		{
			description: "should process a very simple JSON schema",
			inputSchema: fmt.Sprintf("%s/simple-schema.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "simpleSchema",
					Description: "A sample schema for the purpose of testing.",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "string",
							Description: "Sample field description.",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			description: "should process a very simple JSON schema with a required field",
			inputSchema: fmt.Sprintf("%s/simple-schema-required.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "simpleSchema",
					Description: "A sample schema for the purpose of testing.",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "string",
							Required:    true,
							Description: "Sample field description.",
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			description: "should process a JSON schema with multiple fields but no nesting.",
			inputSchema: fmt.Sprintf("%s/one-level-schema.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "oneLevelSchema",
					Description: "A schema with multiple fields but no nested objects.",
					Fields: []Field{
						{
							Name:        "sampleStringField",
							Type:        "string",
							Description: "Sample string field description.",
						},
						{
							Name:        "sampleIntegerField",
							Type:        "integer",
							Description: "Sample integer field description.",
						},
						{
							Name:        "sampleNumberField",
							Type:        "number",
							Description: "Sample number field description.",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			description: "should process a JSON schema with a nested object.",
			inputSchema: fmt.Sprintf("%s/nested-schema.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "nestedSchema",
					Description: "A schema with a single nested object field.",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "string",
							Description: "Sample field description.",
						},
						{
							Name:        "sampleObjectField",
							Type:        "object",
							Description: "Sample object field description.",
						},
					},
				},
				{
					TypeName: "sampleObjectField",
					Fields: []Field{
						{
							Name:        "nestedField",
							Type:        "integer",
							Description: "Nested object field description.",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			description: "should process a JSON schema with a nested object using a definitions ref.",
			inputSchema: fmt.Sprintf("%s/def-schema.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "nestedSchema",
					Description: "A schema with a single nested object field.",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "string",
							Description: "Sample field description.",
						},
						{
							Name:        "sampleObject",
							Type:        "object",
							Description: "Sample object field description.",
						},
					},
				},
				{
					TypeName:    "sampleObject",
					Description: "Sample object field description.",
					Fields: []Field{
						{
							Name:        "nestedField",
							Type:        "integer",
							Description: "Nested object field description.",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			description: "should process a JSON schema with a nested object using a file ref.",
			inputSchema: fmt.Sprintf("%s/def-file-schema.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "nestedSchema",
					Description: "A schema with a single nested object field.",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "string",
							Description: "Sample field description.",
						},
						{
							Name:        "simpleSchema",
							Type:        "object",
							Description: "A sample schema for the purpose of testing.",
						},
					},
				},
				{
					TypeName:    "simpleSchema",
					Description: "A sample schema for the purpose of testing.",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "string",
							Description: "Sample field description.",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			description: "should process a JSON schema with a simple array.",
			inputSchema: fmt.Sprintf("%s/array-schema.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "arraySchema",
					Description: "A schema with a few array fields.",
					Fields: []Field{
						{
							Name:        "arrayStringField",
							Type:        "string",
							Description: "Sample array field description.",
							Array:       true,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			description: "should process a JSON schema with an array of objects.",
			inputSchema: fmt.Sprintf("%s/object-array-schema.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "objectArraySchema",
					Description: "A schema with an array of objects.",
					Fields: []Field{
						{
							Name:        "arrayObjectField",
							Type:        "object",
							Description: "Sample array field description.",
							Array:       true,
						},
						{
							Name:        "secondArrayField",
							Type:        "object",
							Description: "Sample array field description.",
							Array:       true,
						},
					},
				},
				{
					TypeName: "arrayObjectField",
					Fields: []Field{
						{
							Name:        "objectStringField",
							Type:        "string",
							Description: "A string field in an object.",
						},
					},
				},
				{
					TypeName: "secondArrayField",
					Fields: []Field{
						{
							Name:        "objectStringField",
							Type:        "string",
							Description: "A string field in an object.",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			description: "should process a JSON schema with an allOf ref.",
			inputSchema: fmt.Sprintf("%s/schema-with-allOf.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "allOfSchema",
					Description: "A schema with an allOf ref.",
					Fields: []Field{
						{
							Name:        "exampleField",
							Type:        "string",
							Description: "Example field description.",
						},
						{
							Name:        "sampleObjectField",
							Type:        "object",
							Description: "Sample object field description.",
						},
						{
							Name:        "simpleSchema",
							Type:        "object",
							Description: "A sample schema for the purpose of testing.",
						},
					},
				},
				{
					TypeName: "sampleObjectField",
					Fields: []Field{
						{
							Name:        "nestedField",
							Type:        "integer",
							Description: "Nested object field description.",
						},
					},
				},
				{
					TypeName:    "simpleSchema",
					Description: "A sample schema for the purpose of testing.",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "string",
							Description: "Sample field description.",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			description: "should process a JSON schema with a oneOf ref.",
			inputSchema: fmt.Sprintf("%s/schema-with-oneOf.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName:    "oneOfSchema",
					Description: "A schema with a oneOf ref.",
					Fields: []Field{
						{
							Name:        "exampleField",
							Type:        "string",
							Description: "Example field description.",
						},
						{
							Name:        "sampleObjectField",
							Type:        "object",
							Description: "Sample object field description.",
						},
						{
							Name:        "simpleSchema",
							Type:        "object",
							Description: "A sample schema for the purpose of testing.",
						},
					},
				},
				{
					TypeName: "sampleObjectField",
					Fields: []Field{
						{
							Name:        "nestedField",
							Type:        "integer",
							Description: "Nested object field description.",
						},
					},
				},
				{
					TypeName:    "simpleSchema",
					Description: "A sample schema for the purpose of testing.",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "string",
							Description: "Sample field description.",
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			abs, _ := filepath.Abs(test.inputSchema)
			jsonSchema, err := jsonutils.ReadSchema(abs)
			if err != nil {
				t.Fatalf("error reading JSON schema test file at path %q: %v", test.inputSchema, err)
			}

			schemas, err := transform(jsonSchema, abs, jsonSchema.Title)
			// TODO; look at some cleaner error testing.
			if err == nil && test.wantErr != nil {
				t.Errorf("expected the following error, but did not get any error: %v", test.wantErr)
			} else if err != nil {
				if test.wantErr == nil {
					t.Errorf("got the following error when one wasn't expected: %v", err)
				} else if test.wantErr.Error() != err.Error() {
					t.Errorf("did not get the expected error.\nwant- %v\ngot - %v", test.wantErr, err)
				}
			}

			if !reflect.DeepEqual(test.wantGraphQL, schemas) {
				t.Errorf("did not get expected schemas.\nwant - %#v\ngot - %#v", test.wantGraphQL, schemas)
			}
		})
	}
}
