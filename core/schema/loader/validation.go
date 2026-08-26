package loader

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed ctxl.schema.json
var schemaDocument []byte

var (
	compileOnce sync.Once
	compiled    *jsonschema.Schema
	compileErr  error
)

type ValidationError struct {
	Source  string
	Path    string
	Rule    string
	Message string
}

func (e *ValidationError) Error() string {
	prefix := ""
	if e.Source != "" {
		prefix = e.Source + ": "
	}
	where := e.Path
	if where == "" {
		where = "/"
	}
	if e.Rule == "" {
		return fmt.Sprintf("%sschema %s: %s", prefix, where, e.Message)
	}
	return fmt.Sprintf("%sschema %s (%s): %s", prefix, where, e.Rule, e.Message)
}

func validateJSON(raw []byte) error {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return &ValidationError{Rule: "json", Message: err.Error()}
	}
	compileOnce.Do(func() {
		var doc any
		if err := json.Unmarshal(schemaDocument, &doc); err != nil {
			compileErr = err
			return
		}
		compiler := jsonschema.NewCompiler()
		compiler.DefaultDraft(jsonschema.Draft2020)
		if err := compiler.AddResource("https://raw.githubusercontent.com/AkaraChen/ctxl/main/core/schema/loader/ctxl.schema.json", doc); err != nil {
			compileErr = err
			return
		}
		compiled, compileErr = compiler.Compile("https://raw.githubusercontent.com/AkaraChen/ctxl/main/core/schema/loader/ctxl.schema.json")
	})
	if compileErr != nil {
		return fmt.Errorf("compile ctxl json schema: %w", compileErr)
	}
	if err := compiled.Validate(value); err != nil {
		var validation *jsonschema.ValidationError
		if errors.As(err, &validation) {
			return &ValidationError{
				Path:    jsonPointer(validation.InstanceLocation),
				Rule:    jsonPointer(validation.ErrorKind.KeywordPath()),
				Message: validation.Error(),
			}
		}
		return fmt.Errorf("schema validation: %w", err)
	}
	return nil
}

func jsonPointer(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	encoded := make([]string, len(tokens))
	for i, token := range tokens {
		encoded[i] = strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1")
	}
	return "/" + strings.Join(encoded, "/")
}
