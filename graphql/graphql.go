package graphql

import (
	"fmt"
	"jgschema/utils"
	"unicode"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
)

// Schema defines the elements of a GraphQL schema in the context of this program.
type Schema struct {
	TypeName string
	Fields   *[]Field
}

// Field defines the data needed to construct a GraphQL schema field.
type Field struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Array       bool
}

// Transform handles the logic of transforming a given jsonschema.Schema struct into a GraphQL schema struct.
func Transform(jsonSchema *jsonschema.Schema) (*[]Schema, error) {
	return transform(jsonSchema)
}

// TransformFromFile creats a jsonschema.Struct from a file path and transforms it into a GraphQL schema struct.
func TransformFromFile(path string) (*[]Schema, error) {
	schema, err := utils.ReadJSONSchema(path)
	if err != nil {
		return nil, fmt.Errorf("error reading the %q schema defined in allOf block: %w", path, err)
	}

	return transform(schema)
}

func transform(jsonSchema *jsonschema.Schema) (*[]Schema, error) {
	// schemas acts as the "master list" of schemas to be generated.
	var schemas []Schema

	// Handle any allOF schemas defined first.
	if jsonSchema.AllOf != nil {
		for _, allOfSchema := range jsonSchema.AllOf {
			allOfSchemas, err := TransformFromFile(allOfSchema.Ref)
			if err != nil {
				return nil, fmt.Errorf("error processing allOf schema: %w", err)
			}

			schemas = append(schemas, *allOfSchemas...)
		}
	}

	return propertiesWalk(jsonSchema.Properties, &schemas, jsonSchema.Required, jsonSchema.Title)
}

// propertiesWalk walks down the properties tree of a JSON schema, and builds schemas along the way.
func propertiesWalk(root *orderedmap.OrderedMap, schemas *[]Schema, required []string, typeName string) (*[]Schema, error) {
	schema := Schema{
		TypeName: typeName,
		Fields:   &[]Field{},
	}
	for _, key := range root.Keys() {
		property, ok := root.Get(key)
		if !ok {
			continue
		}

		description, err := getOrderedMapKey[string](property, "description")
		if err != nil {
			return nil, err
		}

		fieldType, err := getOrderedMapKey[string](property, "type")
		if err != nil {
			return nil, err
		}

		graphTypeName, err := constructFieldName(key, *fieldType)
		if err != nil {
			return nil, fmt.Errorf("error constructing graphql field name: %w", err)
		}

		*schema.Fields = append(*schema.Fields, Field{
			Name:        key,
			Type:        graphTypeName,
			Description: *description,
			Required:    contains(required, key),
			Array:       isArray(*fieldType),
		})
		schema.TypeName = typeName

		if *fieldType == "object" {
			properties, err := getOrderedMapKey[orderedmap.OrderedMap](property, "properties")
			if err != nil {
				return nil, err
			}

			// Avoid further traversal if there are no properties.
			if properties.Keys() == nil {
				continue
			}

			reqRaw, err := getOrderedMapKey[[]any](property, "required")
			if err != nil {
				return nil, err
			}

			required := make([]string, len(*reqRaw))
			for i, req := range *reqRaw {
				reqStr := req.(string)
				required[i] = reqStr
			}

			return propertiesWalk(properties, schemas, required, title(key))
		}
		*schemas = append(*schemas, schema)
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

func isArray(typeName string) bool {
	return typeName == "array"
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
		return title(name), nil
	default:
		return "", fmt.Errorf("unrecognized type name: %q", typeName)
	}
}

// title uppercases the first letter of a string, per GraphQL's type naming convention.
func title(str string) string {
	r := []rune(str)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
