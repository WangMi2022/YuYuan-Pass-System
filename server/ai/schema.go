package ai

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

func ValidateOutputSchema(schemaText string) error {
	_, err := compileOutputSchema(schemaText)
	return err
}

func validateOutput(schemaText, content string) error {
	if len(content) > maxCompletionOutputBytes {
		return &Error{Type: ErrorTypeSchema, Message: "模型输出超过 4MB 限制"}
	}
	compiled, err := compileOutputSchema(schemaText)
	if err != nil || compiled == nil {
		return err
	}
	var value any
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &value); err != nil {
		return &Error{Type: ErrorTypeSchema, Message: "模型输出不是合法 JSON", Cause: err}
	}
	if err := compiled.Validate(value); err != nil {
		return &Error{Type: ErrorTypeSchema, Message: "模型输出不符合 JSON Schema", Cause: err}
	}
	return nil
}

func compileOutputSchema(schemaText string) (*jsonschema.Schema, error) {
	if strings.TrimSpace(schemaText) == "" {
		return nil, nil
	}
	compiler := jsonschema.NewCompiler()
	compiler.LoadURL = func(string) (io.ReadCloser, error) {
		return nil, errors.New("external JSON Schema references are disabled")
	}
	if err := compiler.AddResource("ai-output-schema.json", strings.NewReader(schemaText)); err != nil {
		return nil, &Error{Type: ErrorTypeSchema, Message: "Prompt 输出 Schema 不合法", Cause: err}
	}
	compiled, err := compiler.Compile("ai-output-schema.json")
	if err != nil {
		return nil, &Error{Type: ErrorTypeSchema, Message: "Prompt 输出 Schema 不合法", Cause: err}
	}
	return compiled, nil
}
