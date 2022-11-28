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

		if schema.Description != "" {
			sb.WriteString(fmt.Sprintf("\"%s\"\n", schema.Description))
		}

		sb.WriteString(fmt.Sprintf("type %s {\n", title(schema.TypeName)))
		for j, field := range schema.Fields {
			if j != 0 && j != len(schema.Fields) && field.Description != "" {
				sb.WriteString("\n")
			}
			if field.Description != "" {
				sb.WriteString(fmt.Sprintf("\t\"%s\"\n", field.Description))
			}
			typeName, err := buildTypeRef(field)
			if err != nil {
				return fmt.Errorf("error building type ref in generate: %w", err)
			}
			sb.WriteString(fmt.Sprintf("\t%s: %s\n", field.Name, typeName))
		}

		sb.WriteString("}")
	}

	_, err := w.Write([]byte(sb.String()))
	return err
}

// buildTypeRef builds the type reference based on the name of the type, whether it is required, and whether it is an array.
func buildTypeRef(field Field) (string, error) {
	builtType, err := constructFieldName(field.Name, field.Type)
	if err != nil {
		return "", fmt.Errorf("error building type reference for field %q: %w", field.Name, err)
	}

	if field.Array {
		builtType = fmt.Sprintf("[%s]", builtType)
	}

	if field.Required {
		builtType = fmt.Sprintf("%s!", builtType)
	}

	return builtType, nil
}

// constructFieldName turns a JSON Schema's type (lowercase) into an uppercase name. If an object, use the name of the field.
func constructFieldName(name string, typeName string) (string, error) {
	switch typeName {
	case "integer":
		return "Int", nil
	case "boolean":
		return "Boolean", nil
	case "number":
		return "Float", nil
	case "string":
		return "String", nil
	case "array":
		return "[]", nil
	default:
		return title(typeName), nil
	}
}
