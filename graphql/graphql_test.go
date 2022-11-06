package graphql

import (
	"fmt"
	"jgschema/jsonutils"
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

	schemaTestDir := "./test_data/schemas"
	tests := []test{
		{
			description: "should process a very simple JSON schema",
			inputSchema: fmt.Sprintf("%s/simple-schema.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName: "SimpleSchema",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "String",
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
					TypeName: "OneLevelSchema",
					Fields: []Field{
						{
							Name:        "sampleStringField",
							Type:        "String",
							Description: "Sample string field description.",
						},
						{
							Name:        "sampleIntegerField",
							Type:        "Int",
							Description: "Sample integer field description.",
						},
						{
							Name:        "sampleNumberField",
							Type:        "Float",
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
					TypeName: "NestedSchema",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "String",
							Description: "Sample field description.",
						},
						{
							Name:        "sampleObjectField",
							Type:        "SampleObjectField",
							Description: "Sample object field description.",
						},
					},
				},
				{
					TypeName: "SampleObjectField",
					Fields: []Field{
						{
							Name:        "nestedField",
							Type:        "Int",
							Description: "Nested object field description.",
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
					TypeName: "AllOfSchema",
					Fields: []Field{
						{
							Name:        "exampleField",
							Type:        "String",
							Description: "Example field description.",
						},
						{
							Name:        "sampleObjectField",
							Type:        "SampleObjectField",
							Description: "Sample object field description.",
						},
						{
							Name:        "simpleSchema",
							Type:        "SimpleSchema",
							Description: "A sample schema for the purpose of testing.",
						},
					},
				},
				{
					TypeName: "SampleObjectField",
					Fields: []Field{
						{
							Name:        "nestedField",
							Type:        "Int",
							Description: "Nested object field description.",
						},
					},
				},
				{
					TypeName: "SimpleSchema",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "String",
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
					TypeName: "AllOfSchema",
					Fields: []Field{
						{
							Name:        "exampleField",
							Type:        "String",
							Description: "Example field description.",
						},
						{
							Name:        "sampleObjectField",
							Type:        "SampleObjectField",
							Description: "Sample object field description.",
						},
					},
				},
				{
					TypeName: "SampleObjectField",
					Fields: []Field{
						{
							Name:        "nestedField",
							Type:        "Int",
							Description: "Nested object field description.",
						},
					},
				},
				{
					TypeName: "SimpleSchema",
					Fields: []Field{
						{
							Name:        "sampleField",
							Type:        "String",
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
			jsonSchema, err := jsonutils.ReadSchema(test.inputSchema)
			if err != nil {
				t.Fatalf("error reading JSON schema test file at path %q: %v", test.inputSchema, err)
			}

			schemas, err := transform(jsonSchema)
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

			if !reflect.DeepEqual(test.wantGraphQL, *schemas) {
				t.Errorf("did not get expected schemas.\nwant - %#v\ngot - %#v", test.wantGraphQL, *schemas)
			}
		})
	}
}
