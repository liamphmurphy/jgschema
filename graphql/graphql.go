package graphql

import (
	"fmt"
	"jgschema/jsonutils"
	"os"
	"path/filepath"
	"strings"
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
func Transform(jsonSchema *jsonschema.Schema, schemaPath string) ([]Schema, error) {
	return transform(jsonSchema, schemaPath)
}

// transform handles the logic of transforming a given jsonschema.Schema struct into a GraphQL schema struct.
func transform(jsonSchema *jsonschema.Schema, schemaPath string) ([]Schema, error) {
	if jsonSchema.Title == "" {
		return nil, fmt.Errorf("please provide a title for the schema")
	}

	parent := Schema{
		TypeName: jsonSchema.Title,
		Fields:   []Field{},
	}

	schemas := []Schema{{}}

	schemaPath = filepath.Dir(schemaPath)

	// To go down the properties tree, we will begin a recursive walk.
	if err := walk(jsonSchema.Properties, &parent, &schemas, typeRoot, jsonSchema.Definitions, schemaPath); err != nil {
		return nil, fmt.Errorf("error when walking down the properties tree: %w", err)
	}

	if jsonSchema.AllOf != nil {
		for _, allOf := range jsonSchema.AllOf {
			err := processRefFile(allOf.Ref, &parent, &schemas, schemaPath)
			if err != nil {
				return nil, fmt.Errorf("error processing allOf schema %q: %w", allOf.Ref, err)
			}
		}
	}

	if jsonSchema.OneOf != nil {
		for _, oneOf := range jsonSchema.OneOf {
			err := processRefFile(oneOf.Ref, &parent, &schemas, schemaPath)
			if err != nil {
				return nil, fmt.Errorf("error processing oneOf schema %q: %w", oneOf.Ref, err)
			}
		}
	}

	if jsonSchema.AnyOf != nil {
		for _, anyOf := range jsonSchema.AllOf {
			err := processRefFile(anyOf.Ref, &parent, &schemas, schemaPath)
			if err != nil {
				return nil, fmt.Errorf("error processing anyOf schema %q: %w", anyOf.Ref, err)
			}
		}
	}

	schemas[0] = parent

	return schemas, nil
}

// walk facilitates the different node types (top of the schema, objects, arrays, etc.) and walks down whatever tree
// that comes from the passed in node.
func walk(node any, parent *Schema, schemas *[]Schema, typeName string, definitions jsonschema.Definitions, schemaPath string) error {
	switch typeName {
	case typeRoot:
		rootOrderedMap, ok := node.(*orderedmap.OrderedMap)
		if !ok {
			return fmt.Errorf("error asserting orderedMap on root node")
		}
		return walkObject(rootOrderedMap, parent, schemas, definitions, schemaPath)
	case typeObject:
		properties, err := extractLeaf(node, "properties")
		if err != nil {
			return fmt.Errorf("error getting properties declaration: %w", err)
		}
		return walkObject(properties, parent, schemas, definitions, schemaPath)
	case typeArray:
		items, err := extractLeaf(node, "items")
		if err != nil {
			return fmt.Errorf("error getting items declaration: %w", err)
		}
		return walkArray(items, parent, schemas, definitions, schemaPath)
	}
	return nil
}

func walkObject(root *orderedmap.OrderedMap, parent *Schema, schemas *[]Schema, definitions jsonschema.Definitions, schemaPath string) error {
	// .Keys() will contain the list of fields from a properties declaration.
	for _, key := range root.Keys() {
		schema := Schema{Fields: []Field{}}
		property, ok := root.Get(key)
		if !ok {
			return fmt.Errorf("property with key %q not found in walkObject", key)
		}

		schema.TypeName = key

		potentialRef, _ := getOrderedMapKey[string](property, "$ref")
		if *potentialRef != "" {
			var definition *jsonschema.Schema

			refPath := parseRefPath(*potentialRef)
			// Determine the ref path for grabbing the definition; either from a $defs path in the current schema, or a separate file.
			if refPath != "" {
				definition = definitions[refPath]
				if definition == nil {
					// If the reference to the other schema is not absolute, we'll need to build an absolute path version to ensure correctness.
					if !filepath.IsAbs(refPath) {
						refPath = filepath.Clean(fmt.Sprintf("%s/%s", schemaPath, refPath))
					}
					if _, err := os.Stat(refPath); err == nil {
						definition, err = jsonutils.ReadSchema(refPath)
						if err != nil {
							return fmt.Errorf("error reading ref file: %w", err)
						}
					} else {
						return fmt.Errorf("provided ref %q was not a valid definition or external schema file: %w", refPath, err)
					}
				}
			} else {
				return fmt.Errorf("received invalid ref: %q", refPath)
			}

			if err := processRef(definition, &schema, schemas, schemaPath); err != nil {
				return fmt.Errorf("error processing ref at %q", key)
			}

			parent.Fields = append(parent.Fields, schema.Fields...)
			return nil
		}

		// Ignore error on description as it isn't required to build the GraphQL schema.
		description, _ := getOrderedMapKey[string](property, "description")
		if description == nil {
			var blank string
			description = &blank
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

		// Declare the field early and let any further traversal operations update the field if needed.
		field := Field{
			Name:        key,
			Description: *description,
			Type:        formattedFieldType,
			Array:       isArray(*fieldType),
		}

		switch *fieldType {
		case typeObject:
			schema.TypeName = title(key)

			if err := walk(property, &schema, schemas, typeObject, definitions, schemaPath); err != nil {
				return fmt.Errorf("error walking down nested object %q: %w", key, err)
			}

			*schemas = append(*schemas, schema)
		case typeArray:
			if err := walk(property, &schema, schemas, typeArray, definitions, schemaPath); err != nil {
				return fmt.Errorf("error walking down array %q: %w", key, err)
			}

			// TODO: don't assume index 0 for the field
			if schema.Fields != nil {
				field.Type = schema.Fields[0].Type
			}
		}

		parent.Fields = append(parent.Fields, field)

	}

	return nil
}

func walkArray(root *orderedmap.OrderedMap, parent *Schema, schemas *[]Schema, definitions jsonschema.Definitions, schemaPath string) error {
	// .Keys() will contain the list of fields from an items declaration.
	for _, key := range root.Keys() {
		raw, ok := root.Get(key)
		if !ok {
			return fmt.Errorf("key value not found")
		}
		switch key {
		case "type":

			fieldType := raw.(string)

			if fieldType == typeObject {
				// TODO: re-use code from here and the default case.
				newSchema := Schema{
					TypeName: title(parent.TypeName),
					Fields:   []Field{},
				}
				if err := walk(root, &newSchema, schemas, typeObject, definitions, schemaPath); err != nil {
					return fmt.Errorf("error walking down object array item %q: %w", key, err)
				}

				// Merge newSchema back into the parent.
				parent.Fields = append(parent.Fields, newSchema.Fields...)
				parent.Fields[0].Type = fmt.Sprintf("[%s]", title(parent.TypeName))
				*schemas = append(*schemas, newSchema)

				return nil
			}

			// Depending on the type, the casing can change (mainly with objects), so some extra formatting is needed.
			formattedFieldType, err := constructFieldName(key, fieldType)
			if err != nil {
				return fmt.Errorf("error constructing field type: %w", err)
			}
			field := Field{
				Type: fmt.Sprintf("[%s]", formattedFieldType),
			}

			parent.Fields = append(parent.Fields, field)
		default:
			// This key could be an object, so entertain that first before erroring on an unknown key.
			properties, err := extractLeaf(raw, "properties")
			if err != nil {
				return fmt.Errorf("unknown key in array items: %q", key)
			}
			newSchema := Schema{
				TypeName: title(parent.TypeName),
				Fields:   []Field{},
			}
			if err = walkObject(properties, &newSchema, schemas, definitions, schemaPath); err != nil {
				return fmt.Errorf("error walking down object array item %q: %w", key, err)
			}

			// Merge newSchema back into the parent.
			parent.Fields = append(parent.Fields, newSchema.Fields...)
			parent.Fields[0].Type = fmt.Sprintf("[%s]", title(parent.TypeName))
			*schemas = append(*schemas, newSchema)
		}
	}

	return nil
}

func processRefFile(refPath string, parent *Schema, schemas *[]Schema, schemaPath string) error {
	refSchema, err := jsonutils.ReadSchema(refPath)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	return processRef(refSchema, parent, schemas, schemaPath)
}

// processRef generalizes the logic for processing allOf, oneOf, and anyOf refs.
// Since walk isn't smart enough to know when a ref is being passed down, we manually
// append the results of the walk to the parent (root) and schemas list.
func processRef(schema *jsonschema.Schema, parent *Schema, schemas *[]Schema, schemaPath string) error {
	var refGraphQL Schema
	if err := walk(schema.Properties, &refGraphQL, schemas, typeRoot, schema.Definitions, schemaPath); err != nil {
		return fmt.Errorf("error processing allOf schema %q: %w", schema.Title, err)
	}

	// Some refs won't contain titles, in that case borrow from the parent.
	if schema.Title == "" {
		schema.Title = title(parent.TypeName)
	}

	parent.Fields = append(parent.Fields, Field{
		Name:        lowerTitle(schema.Title),
		Description: schema.Description,
		Type:        schema.Title,
	})
	refGraphQL.TypeName = schema.Title
	*schemas = append(*schemas, refGraphQL)

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

// parseRefPath splits a "$ref" of the style "#/$defs/fieldName" for finding the definition name.
func parseRefPath(path string) string {
	split := strings.Split(path, "/")
	if len(split) == 0 {
		return ""
	}

	return split[len(split)-1]
}
