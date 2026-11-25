package ai

import "testing"

func TestValidateOutputSchema(t *testing.T) {
	schema := `{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`
	if err := ValidateOutputSchema(schema); err != nil {
		t.Fatal(err)
	}
	if err := validateOutput(schema, `{"name":"打印机"}`); err != nil {
		t.Fatal(err)
	}
	if err := validateOutput(schema, `{"name":12}`); ErrorKind(err) != ErrorTypeSchema {
		t.Fatalf("error kind = %q, want %q", ErrorKind(err), ErrorTypeSchema)
	}
	if err := ValidateOutputSchema(`{"type":`); ErrorKind(err) != ErrorTypeSchema {
		t.Fatalf("error kind = %q, want %q", ErrorKind(err), ErrorTypeSchema)
	}
	if err := ValidateOutputSchema(`{"$ref":"http://127.0.0.1/schema.json"}`); ErrorKind(err) != ErrorTypeSchema {
		t.Fatalf("remote ref error kind = %q, want %q", ErrorKind(err), ErrorTypeSchema)
	}
}
