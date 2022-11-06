package graphql

import (
	"fmt"
	"os"
	"reflect"
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
				Schema{
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
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			generated, err := Generate(test.inputGraphQL)
			if err != nil {
				t.Fatalf("error reading JSON schema test file at path %q: %v", test.wantSchema, err)
			}

			want, err := os.ReadFile(test.wantSchema)
			if err != nil {
				t.Fatalf("error reading test graphql schema file at path %q: %v", test.wantSchema, err)
			}

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

			if !reflect.DeepEqual(string(want), generated) {
				t.Errorf("did not get expected generated result.\nwant - %s\ngot - %s", want, generated)
			}
		})
	}
}
