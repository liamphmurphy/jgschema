package graphql

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	type test struct {
		description  string
		inputGraphQL []Schema
		wantSchema   string // path to test data file
		wantErr      error
	}

	schemaTestDir := "./test_data/graphschema"
	tests := []test{
		{
			description: "Should successfully generate from a simple graphql schema.",
			inputGraphQL: []Schema{
				{
					TypeName: "Test",
					Fields: []Field{
						{
							Name:        "testField",
							Description: "Test description.",
							Type:        "String",
						},
					},
				},
			},
			wantSchema: fmt.Sprintf("%s/simple-schema.graphql", schemaTestDir),
		},
		{
			description: "Should successfully generate from a simple schema with an object field.",
			inputGraphQL: []Schema{
				{
					TypeName: "Test",
					Fields: []Field{
						{
							Name:        "testField",
							Description: "Test description.",
							Type:        "String",
						},
						{
							Name:        "testObject",
							Description: "Test object.",
							Type:        "TestObject",
						},
					},
				},
				{
					TypeName: "TestObject",
					Fields: []Field{
						{
							Name:        "objectField",
							Description: "object field description.",
							Type:        "Int",
						},
					},
				},
			},
			wantSchema: fmt.Sprintf("%s/object-schema.graphql", schemaTestDir),
		},
		{
			description: "Should successfully generate from a simple schema with array and required fields.",
			inputGraphQL: []Schema{
				{
					TypeName: "Test",
					Fields: []Field{
						{
							Name:        "testField",
							Description: "Test description.",
							Type:        "String",
							Array:       true,
						},
						{
							Name:        "testObject",
							Description: "Test object.",
							Type:        "TestObject",
							Required:    true,
						},
					},
				},
				{
					TypeName: "TestObject",
					Fields: []Field{
						{
							Name:        "objectField",
							Description: "object field description.",
							Type:        "Int",
						},
					},
				},
			},
			wantSchema: fmt.Sprintf("%s/array-required-fields.graphql", schemaTestDir),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			generated, err := Generate(test.inputGraphQL)
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

			want, err := os.ReadFile(test.wantSchema)
			if err != nil {
				t.Fatalf("error reading test graphql schema file at path %q: %v", test.wantSchema, err)
			}

			cleanedWant := cleanUpFileContents(string(want))
			cleanedGenerated := cleanUpFileContents(generated)

			if cleanedWant != cleanedGenerated {
				t.Errorf("did not get expected generated result.\nwant - %s\ngot - %s", want, generated)
			}
		})
	}
}

func cleanUpFileContents(contents string) string {
	for _, char := range []string{"\n", "\t", " "} {
		contents = strings.ReplaceAll(contents, char, "")
	}
	return contents
}
