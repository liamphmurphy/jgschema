package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
)

func main() {
	example, err := os.ReadFile("./example.json")
	if err != nil {
		fmt.Printf("error reading example file: %v\n", err)
		os.Exit(1)
	}

	var schema jsonschema.Schema
	if err := json.Unmarshal(example, &schema); err != nil {
		fmt.Printf("error unmarshaling to json schema: %v\n", err)
		os.Exit(1)
	}

	graphSchema, err := Transform(&schema)
	if err != nil {
		fmt.Printf("error transforming graphql schema: %v", err)
	}

	fmt.Printf("%#v\n", *graphSchema)
}
