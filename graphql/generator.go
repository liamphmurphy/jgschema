package graphql

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"text/template"
)

type Generator struct {
	Schemas []Schema
}

var (
	schemaTemplate = "./graphql/schema.tmpl"
)

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

// generate contains the core template logic for generating a GraphQL schema list. template.Execute makes use of an io.Writer,
// so a passed in Writer is needed and the resulting template will be allocated to the Writer.
func generate(schemas []Schema, writer io.Writer) error {
	if writer == nil {
		return fmt.Errorf("an io.Writer was not passed into the generator")
	}

	generator := Generator{Schemas: schemas}

	t, err := template.ParseFiles(schemaTemplate)
	if err != nil {
		return fmt.Errorf("error creating template: %w", err)
	}

	if err := t.Execute(writer, &generator); err != nil {
		return fmt.Errorf("error executing graphql template: %w", err)
	}

	return nil
}
