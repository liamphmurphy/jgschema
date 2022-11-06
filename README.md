jgschema (working name!) is a small CLI tool written in Go to convert JSON schemas to GraphQL schemas.

This tool helps eliminate some of the tedium of maintaining a JSON schema and GraphQL schema for data models that begin with the creation of a JSON schema. Professionally I have a case where we define a datatype model starting with a JSON schema, generate code models (Go), and then support the querying of that datatype in a GraphQL API.

A few issues arise when you have to manually translate a JSON schema to a GraphQL schema:
- Schema drift as people forget to update the GraphQL schema after a JSON schema change.
- Manual translation can cause typos, not just in type names but even something as small as a description. 
- Not obvious to new developers on the app that this manual translation is a part of the flow. 
- It takes a lot of time. 

# What this app does not do
- Does not support taking in a JSON payload (non-schema) and turning that into GraphQL.
- Does not support the reverse operation of translating a GraphQL schema to a JSON schema.

# Features
Below are the list of features that are either done or need to be worked on.

✅ Translates the following JSON types: scalars (strings, integers, numbers, boolean) and objects.
x Support allOf in any place in the properties tree.
x Support arrays.
x Generator for writing the GraphQL file.
x support oneOf's / anyOf's..