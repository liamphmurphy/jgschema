package main

import (
	"fmt"
	"unicode"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
)

// Schema defines the elements of a GraphQL schema in the context of this program.
type Schema struct {
	TypeName string
	Fields   []Field
}

type Field struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

// Transform contains all the logic for transforming a JSON schema into a GraphQL schema struct.
func Transform(jsonSchema *jsonschema.Schema) (*[]Schema, error) {
	var schema Schema
	schema.TypeName = jsonSchema.Title

	var schemas []Schema

	return propertiesWalk(jsonSchema.Properties, &schemas, jsonSchema.Required)
}

// propertiesWalk walks down the properties tree of a JSON schema, and builds schemas along the way.
func propertiesWalk(root *orderedmap.OrderedMap, schemas *[]Schema, required []string) (*[]Schema, error) {
	for _, key := range root.Keys() {
		var fields []Field

		get, ok := root.Get(key)
		if !ok {
			continue
		}

		description, err := getOrderedMapKey[string](get, "description")
		if err != nil {
			return nil, err
		}

		fieldType, err := getOrderedMapKey[string](get, "type")
		if err != nil {
			return nil, err
		}

		graphTypeName, err := constructFieldName(key, *fieldType)
		if err != nil {
			return nil, fmt.Errorf("error constructing graphql field name: %w", err)
		}

		fields = append(fields, Field{
			Name:        key,
			Type:        graphTypeName,
			Description: *description,
			Required:    contains(required, key),
		})

		*schemas = append(*schemas, Schema{TypeName: *fieldType, Fields: fields})

		if *fieldType == "object" {
			properties, err := getOrderedMapKey[orderedmap.OrderedMap](get, "properties")
			if err != nil {
				return nil, err
			}

			// Avoid recursion if there are no further properties to process.
			if len(properties.Keys()) == 0 {
				return schemas, nil
			}

			reqRaw, err := getOrderedMapKey[[]any](get, "required")
			if err != nil {
				return nil, err
			}

			required := make([]string, len(*reqRaw))
			for i, req := range *reqRaw {
				reqStr := req.(string)
				required[i] = reqStr
			}

			return propertiesWalk(properties, schemas, required)
		}

	}

	return schemas, nil
}

// assertOrderedMapValue generically automates the tedium of getting a value out of an orderedmap.OrderedMap key/value pair.
func getOrderedMapKey[T any](property any, key string) (*T, error) {
	orderedMap, ok := property.(orderedmap.OrderedMap)
	if !ok {
		return new(T), fmt.Errorf("error asserting that property of type %T is a %T", orderedMap, orderedmap.OrderedMap{})
	}

	value, ok := orderedMap.Get(key)
	if !ok {
		return new(T), nil
	}

	assertion, ok := value.(T)
	if !ok {
		return &assertion, fmt.Errorf("the value %v is not a %T", value, new(T))
	}

	return &assertion, nil
}

func contains(requiredFields []string, field string) bool {
	for _, req := range requiredFields {
		if req == field {
			return true
		}
	}
	return false
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
	case "object":
		r := []rune(name)
		r[0] = unicode.ToUpper(r[0])
		return string(r), nil
	default:
		return "", fmt.Errorf("unrecognized type name: %q", typeName)
	}
}
