package graphql

import (
	"errors"
	"fmt"
	"jgschema/jsonutils"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/invopop/jsonschema"
)

func getRef(path, defKeyword, schemaPath string, definitions jsonschema.Definitions) (*jsonschema.Schema, error) {
	if path == "" {
		return nil, errors.New("passed in path to parseRefPath was empty")
	} else if defKeyword == "" {
		return nil, errors.New("passed in defintition keyword to parseRefPath was empty")
	}

	var isDefinition, isExternal bool

	if strings.Contains(path, ".json") {
		isExternal = true
	}

	if strings.Contains(path, fmt.Sprintf("#/%s", defKeyword)) {
		isDefinition = true
	}

	var schema *jsonschema.Schema
	switch {
	case isExternal && isDefinition:

	case isExternal:
		if !filepath.IsAbs(path) {
			split := strings.Split(path, "/")
			path = filepath.Clean(fmt.Sprintf("%s/%s", schemaPath, split[len(split)-1]))
		}
		return jsonutils.ReadSchema(path)
	case isDefinition:
		split := strings.Split(path, "/")
		definitionName := split[len(split)-1]

		schema = definitions[definitionName]
		if schema == nil {
			return nil, fmt.Errorf("there is no definition named %q in this schema", definitionName)
		}

		if schema.Title == "" {
			schema.Title = definitionName
		}
	}

	return schema, nil
}

// walkRef generalizes the logic for processing allOf, oneOf, and anyOf refs.
// Since walk isn't smart enough to know when a ref is being passed down, we manually
// append the results of the walk to the parent (root) and schemas list.
func walkRef(schema *jsonschema.Schema, parent *Schema, schemas *[]Schema, schemaPath string) error {
	refGraphQL := Schema{TypeName: schema.Title, Description: schema.Description}
	if err := walk(schema.Properties, schema.Required, &refGraphQL, schemas, typeRoot, schema.Definitions, schemaPath); err != nil {
		return fmt.Errorf("error processing ref schema %q: %w", schema.Title, err)
	}

	// Some refs won't contain titles, in that case borrow from the parent.
	if schema.Title == "" {
		schema.Title = parent.TypeName
	}

	parent.Fields = append(parent.Fields, Field{
		Name:        lowerTitle(schema.Title),
		Description: schema.Description,
		Type:        typeObject,
	})
	*schemas = append(*schemas, refGraphQL)

	return nil
}

func fileNameNoExtension(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

// title uppercases the first letter of a string, per GraphQL's type naming convention.
func title(str string) string {
	if str == "" {
		return str
	}

	r := []rune(str)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// title lowercases the first letter of a string, per GraphQL's field naming convention.
func lowerTitle(str string) string {
	if str == "" {
		return str
	}

	r := []rune(str)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
