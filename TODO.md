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


- Fix definition ref parse to use actual field name instead of lower case type name as field name, see definitions ref test 
