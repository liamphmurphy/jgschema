package main

import (
	"fmt"
	"jgschema/graphql"
	"jgschema/utils"
	"os"
)

func main() {
	jsonSchema, err := utils.ReadJSONSchema("./example.json")
	if err != nil {
		fmt.Printf("error reading JSON schema: %v\n", err)
		os.Exit(1)
	}

	graphSchema, err := graphql.Transform(jsonSchema)
	if err != nil {
		fmt.Printf("error transforming graphql schema: %v", err)
	}

	fmt.Printf("%#v\n", *graphSchema)
}
