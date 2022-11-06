package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
)

func ReadJSONSchema(path string) (*jsonschema.Schema, error) {
	example, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading example file: %w", err)
	}

	var schema jsonschema.Schema
	if err := json.Unmarshal(example, &schema); err != nil {
		return nil, fmt.Errorf("error unmarshaling to json schema: %w", err)
	}

	return &schema, nil
}
