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
					Fields: &[]Field{
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
		/*{
			description: "should process a JSON schema with multiple fields but no nesting.",
			inputSchema: fmt.Sprintf("%s/one-level-schema.json", schemaTestDir),
			wantGraphQL: []Schema{
				{
					TypeName: "OneLevelSchema",
					Fields: &[]Field{
						{
							Name:        "sampleStringField",
							Type:        "String",
							Description: "Sample field string description.",
						},
						{
							Name:        "sampleIntegerField",
							Type:        "Int",
							Description: "Sample integer string description.",
						},
						{
							Name:        "sampleNumberField",
							Type:        "Float",
							Description: "Sample number string description.",
						},
					},
				},
			},
			wantErr: nil,
		},*/
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