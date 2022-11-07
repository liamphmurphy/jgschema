jgschema (working name!) is a small CLI tool written in Go to convert JSON schemas to GraphQL schemas.

This tool helps eliminate some of the tedium of maintaining a JSON schema and GraphQL schema for data models that begin with the creation of a JSON schema. 

A few issues arise when you have to manually translate a JSON schema to a GraphQL schema:
- Schema drift as people forget to update the GraphQL schema after a JSON schema change.
- Manual translation can cause typos, not just in type names but even something as small as a description. 
- Not obvious to new developers on the app that this manual translation is a part of the flow. 
- It takes a lot of time. 

# Logic Explanation

This tool uses a recursive approach of starting from a "parent schema" and walking down any allOf schemas and the parent schema's properties tree. 

GraphQL does not have the notion of nesting, thus all nested schemas in the recursion are made as a separate type and everything is flattened to the top-level in the schema.

The parent schema will contain fields referencing the first-level of the properties tree; including arrays and objects. 

# What this app does not do
- Does not support taking in a JSON payload (non-schema) and turning that into GraphQL.
- Does not support the reverse operation of translating a GraphQL schema to a JSON schema.
- Does not (and technically cannot) enforce any "valid value restrictions" designated in the JSON schema, such as minLength, maxLength, maxItems, etc. That is up to your GraphQL resolver logic to enforce.

# Features
Below are the list of features that are either done or need to be worked on.

- ✅ Translates the following JSON types: scalars (strings, integers, numbers, boolean) and objects.
- ✅ Support allOf in any place in the properties tree.
- ✅ GraphQL file generator.
- Support arrays.
- Support definitions, both file and inline.
- Support running from Docker.