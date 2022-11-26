package main

import (
	"fmt"
	"jgschema/graphql"
	"jgschema/jsonutils"
	"os"
)

func main() {
	schemaPath := "./example.json"
	jsonSchema, err := jsonutils.ReadSchema(schemaPath)
	if err != nil {
		fmt.Printf("error reading JSON schema: %v\n", err)
		os.Exit(1)
	}

	graphSchema, err := graphql.Transform(jsonSchema, schemaPath)
	if err != nil {
		fmt.Printf("error transforming graphql schema: %v", err)
		os.Exit(1)
	}

	generateSchema, err := graphql.Generate(graphSchema)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(generateSchema)
}
