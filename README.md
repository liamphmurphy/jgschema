This tool helps eliminate some of the tedium of maintaining a JSON schema and GraphQL schema for data models that begin with the creation of a JSON schema. Professionally I have a case where we define a datatype model starting with a JSON schema, generate code models (Go), and then support the querying of that datatype in a GraphQL API.

A few issues arise when you have to manually translate a JSON schema to a GraphQL schema:
- Schema drift as people forget to update the GraphQL schema after a JSON schema change
- Manual translation can cause typos, not just in type names but even something as small as a description. 
- Not obvious to new developers on the app that this manual translation is a part of the flow. 
- It takes a lot of time. 

# What this app does not do
- Does not support taking in a JSON payload (non-schema) and turning that into GraphQL.
- Does not support the reverse operation of translating a GraphQL schema to a JSON schema.

# Features
Below are the list of features that are either done or need to be worked on.

✅ Supports the following JSON types: scalars (strings, integers, numbers, boolean) and objects.
✅ Walk down the "properties" tree for nested object traversal.
x Support allOf in any place in the properties tree.
x Support arrays.

# Tentative plans

- It would be nice to support "oneOf", the issue here is that this case can be a little ambiguous. Current tentative plan is to just support the generation of any schema listed in a "oneOf", by generating the GraphQL schemas for all oneOf's listed but to not to try and do anything fancy by making a GraphQL interface. 