// Package schemajson contains the JSON decoding and validation code generated
// from ctxl's canonical JSON Schema.
package schemajson

//go:generate go run github.com/atombender/go-jsonschema@v0.24.1 -p schemajson --tags json --schema-root-type=https://ctxl.dev/schema/v1.json=Document -o generated.go ../../loader/ctxl.schema.json
