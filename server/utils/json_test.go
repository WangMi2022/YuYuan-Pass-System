package utils

import "testing"

func TestGetJSONKeys(t *testing.T) {
	var jsonStr = `
	{
		"Name": "test",
		"TableName": "test",
		"TemplateID": "test",
		"TemplateInfo": "test",
		"Limit": 0
}`
	keys, err := GetJSONKeys(jsonStr)
	if err != nil {
		t.Fatalf("GetJSONKeys returned an error: %v", err)
	}
	want := []string{"Name", "TableName", "TemplateID", "TemplateInfo", "Limit"}
	if len(keys) != len(want) {
		t.Fatalf("GetJSONKeys returned %d keys, want %d", len(keys), len(want))
	}
	for index, expected := range want {
		if keys[index] != expected {
			t.Errorf("key %d = %q, want %q", index, keys[index], expected)
		}
	}
}
