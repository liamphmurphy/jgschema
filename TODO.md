# This file will be for keeping track of things to do, but don't want to muddle up the README. It'll be deleted at some point.

## Remaining Definitions Work

- Referring to a definition in another file, example:
```
{
    ...
    "$ref": "./schemas/some-other-schema.json#/$defs/fieldName"
    ...
}
```

Related to above, what we do have working:
```
    "$ref": "./schemas/some-other-schema.json" // refer to another schema
    "$ref: "#/$defs/fieldName" // refers to definition in original schema
```

- Supporting `$ref` in an array items type declaration, example:
```
    "thisIsAnArray": {
        "type": "array",
        "items": {
            "$ref": "./some-other-schema.json"
        }
    }
```


1. x Clean up graphql.go deciding how text appears in the graphql schema
2. x In the GQL generated schema, if there's no description / comment, don't add a newline between the next field.
3. x Comment on top of a GQL schema type declaration, if the JSON Schema has a top-level description? 
4. Add support for referirng to definition in an external schema, see above
