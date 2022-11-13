package graphql

import (
	"fmt"
	"jgschema/jsonutils"
	"unicode"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
)

const (
	typeObject = "object"
	typeRoot   = "root"
	typeArray  = "array"
)

// Schema defines the elements of a GraphQL schema in the context of this program.
type Schema struct {
	TypeName string
	Fields   []Field
}

// Field defines the data needed to construct a GraphQL schema field.
type Field struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Array       bool
}

// Transform is a public wrapper around transform, where an already made jsonschema.Schema is used.
func Transform(jsonSchema *jsonschema.Schema) ([]Schema, error) {
	return transform(jsonSchema)
}

// transform handles the logic of transforming a given jsonschema.Schema struct into a GraphQL schema struct.
func transform(jsonSchema *jsonschema.Schema) ([]Schema, error) {
	if jsonSchema.Title == "" {
		return nil, fmt.Errorf("please provide a title for the schema")
	}

	parent := Schema{
		TypeName: jsonSchema.Title,
		Fields:   []Field{},
	}

	schemas := []Schema{{}}

	// To go down the properties tree, we will begin a recursive walk.
	if err := walk(jsonSchema.Properties, &parent, &schemas, typeRoot); err != nil {
		return nil, fmt.Errorf("error when walking down the properties tree: %w", err)
	}

	if jsonSchema.AllOf != nil {
		for _, allOf := range jsonSchema.AllOf {
			err := processRef(allOf.Ref, &parent, &schemas)
			if err != nil {
				return nil, fmt.Errorf("error processing allOf schema %q: %w", allOf.Ref, err)
			}
		}
	}

	if jsonSchema.OneOf != nil {
		for _, oneOf := range jsonSchema.OneOf {
			err := processRef(oneOf.Ref, &parent, &schemas)
			if err != nil {
				return nil, fmt.Errorf("error processing oneOf schema %q: %w", oneOf.Ref, err)
			}
		}
	}

	schemas[0] = parent

	return schemas, nil
}

// walk facilitates the different node types (top of the schema, objects, arrays, etc.) and walks down whatever tree
// that comes from the passed in node.
func walk(node any, parent *Schema, schemas *[]Schema, typeName string) error {
	switch typeName {
	case typeRoot:
		rootOrderedMap, ok := node.(*orderedmap.OrderedMap)
		if !ok {
			return fmt.Errorf("error asserting orderedMap on root node")
		}
		return walkObject(rootOrderedMap, parent, schemas)
	case typeObject:
		properties, err := extractLeaf(node, "properties")
		if err != nil {
			return fmt.Errorf("error getting properties declaration: %w", err)
		}
		return walkObject(properties, parent, schemas)
	case typeArray:
		items, err := extractLeaf(node, "items")
		if err != nil {
			return fmt.Errorf("error getting items declaration: %w", err)
		}
		return walkArray(items, parent)
	}
	return nil
}

func walkObject(root *orderedmap.OrderedMap, parent *Schema, schemas *[]Schema) error {
	// .Keys() will contain the list of fields from a properties declaration.
	for _, key := range root.Keys() {
		schema := Schema{Fields: []Field{}}
		property, ok := root.Get(key)
		if !ok {
			return fmt.Errorf("property with key %q not found in walkObject", key)
		}

		schema.TypeName = key

		// Ignore error on description as it isn't required to build the GraphQL schema.
		description, _ := getOrderedMapKey[string](property, "description")
		if description == nil {
			*description = ""
		}

		fieldType, err := getOrderedMapKey[string](property, "type")
		if err != nil {
			return fmt.Errorf("error on field %q getting object field type: %w", key, err)
		}

		// Depending on the type, the casing can change (mainly with objects), so some extra formatting is needed.
		formattedFieldType, err := constructFieldName(key, *fieldType)
		if err != nil {
			return fmt.Errorf("error constructing field type: %w", err)
		}

		field := Field{
			Name:        key,
			Description: *description,
			Type:        formattedFieldType,
			Array:       isArray(*fieldType),
		}

		if *fieldType == typeObject {
			schema.TypeName = title(key)

			if err := walk(property, &schema, schemas, typeObject); err != nil {
				return fmt.Errorf("error walking down nested object %q: %w", key, err)
			}

			*schemas = append(*schemas, schema)
		} else if *fieldType == typeArray {
			if err := walk(property, &schema, schemas, typeArray); err != nil {
				return fmt.Errorf("error walking down array %q: %w", key, err)
			}
		}
		parent.Fields = append(parent.Fields, field)

	}

	return nil
}

func walkArray(root *orderedmap.OrderedMap, parent *Schema) error {
	// .Keys() will contain the list of fields from an items declaration.
	for _, key := range root.Keys() {
		switch key {
		case "type":
			typeRaw, ok := root.Get(key)
			if !ok {
				return fmt.Errorf("type value not found")
			}

			fieldType := typeRaw.(string)

			// Depending on the type, the casing can change (mainly with objects), so some extra formatting is needed.
			formattedFieldType, err := constructFieldName(key, fieldType)
			if err != nil {
				return fmt.Errorf("error constructing field type: %w", err)
			}
			field := Field{
				Type: fmt.Sprintf("[%s]", formattedFieldType),
			}

			parent.Fields = append(parent.Fields, field)
		}
	}

	return nil
}

// processRef generalizes the logic for processing allOf, oneOf, and anyOf refs.
// Since walk isn't smart enough to know when a ref is being passed down, we manually
// append the results of the walk to the parent (root) and schemas list.
func processRef(refPath string, parent *Schema, schemas *[]Schema) error {
	refSchema, err := jsonutils.ReadSchema(refPath)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	var allOfGraphQL Schema
	if err := walk(refSchema.Properties, &allOfGraphQL, schemas, typeRoot); err != nil {
		return fmt.Errorf("error processing allOf schema %q: %w", refSchema.Title, err)
	}

	parent.Fields = append(parent.Fields, Field{
		Name:        lowerTitle(refSchema.Title),
		Description: refSchema.Description,
		Type:        refSchema.Title,
	})
	allOfGraphQL.TypeName = refSchema.Title
	*schemas = append(*schemas, allOfGraphQL)

	return nil
}

func extractLeaf(node any, key string) (*orderedmap.OrderedMap, error) {
	orderedMap, err := getOrderedMapKey[orderedmap.OrderedMap](node, key)
	if err != nil {
		return nil, fmt.Errorf("error extracting %q from node: %w", key, err)
	}

	return orderedMap, nil
}

// assertOrderedMapValue generically automates the tedium of getting a value out of an orderedmap.OrderedMap key/value pair.
func getOrderedMapKey[T any](property any, key string) (*T, error) {
	orderedMap, ok := property.(orderedmap.OrderedMap)
	if !ok {
		retry, ok := property.(*orderedmap.OrderedMap)
		if !ok {
			return new(T), fmt.Errorf("error asserting that property of type %T is a %T", orderedMap, orderedmap.OrderedMap{})
		} else {
			orderedMap = *retry
		}
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
	case "array":
		return "[]", nil
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

// title lowercases the first letter of a string, per GraphQL's field naming convention.
func lowerTitle(str string) string {
	r := []rune(str)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
