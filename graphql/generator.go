package graphql

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

type Generator struct {
	Schemas []Schema
}

// Generate will return the full GraphQL generated schema as a string.
func Generate(schemas []Schema) (string, error) {
	var buffer bytes.Buffer

	if err := generate(schemas, &buffer); err != nil {
		return "", fmt.Errorf("error generating graphql schema: %w", err)
	}

	return buffer.String(), nil
}

// GenerateToFile will write the full GraphQL generated schema into a file with the passed in path and permissions.
func GenerateToFile(schemas []Schema, path string, perms fs.FileMode) error {
	var buffer bytes.Buffer
	if err := generate(schemas, &buffer); err != nil {
		return fmt.Errorf("error generating graphql schema: %w", err)
	}

	return os.WriteFile(path, buffer.Bytes(), perms)
}

// generate takes in a slice of GraphQL schemas and writes it in the format of a GraphQL schema file.
// Final result is allocated to the passed in io.Writer.
func generate(schemas []Schema, w io.Writer) error {
	if w == nil {
		return fmt.Errorf("an io.Writer was not passed into the generator")
	}

	var sb strings.Builder

	for i, schema := range schemas {
		if i != 0 && i != len(schemas) {
			sb.WriteString("\n\n")
		}

		sb.WriteString(fmt.Sprintf("type %s {\n", schema.TypeName))
		for j, field := range schema.Fields {
			if j != 0 && j != len(schema.Fields) {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("\t\"%s\"\n", field.Description))
			sb.WriteString(fmt.Sprintf("\t%s: %s\n", field.Name, buildTypeRef(field)))
		}

		sb.WriteString("}")
	}

	_, err := w.Write([]byte(sb.String()))
	return err
}

// buildTypeRef builds the type reference based on the name of the type, whether it is required, and whether it is an array.
func buildTypeRef(field Field) string {
	builtType := field.Type
	if field.Array {
		builtType = fmt.Sprintf("[%s]", builtType)
	}

	if field.Required {
		builtType = fmt.Sprintf("%s!", builtType)
	}

	return builtType
}
