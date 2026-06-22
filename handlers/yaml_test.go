package handlers

import (
	"strings"
	"testing"
)

func TestJSONToYAMLLogic(t *testing.T) {
	jsonInput := `{"z": {"c": "nested_first", "a": "nested_second"}, "a": "second"}`
	res, err := JSONToYAMLLogic(jsonInput, "4", "single")
	if err != nil {
		t.Fatalf("JSONToYAMLLogic error: %v", err)
	}

	expected := `z:
    c: nested_first
    a: nested_second
a: second
`
	if strings.TrimSpace(res) != strings.TrimSpace(expected) {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, res)
	}
}

func TestYAMLToJSONLogic(t *testing.T) {
	yamlInput := `
z:
  c: nested_first
  a: nested_second
a: second
`
	res, err := YAMLToJSONLogic(yamlInput, "2")
	if err != nil {
		t.Fatalf("YAMLToJSONLogic error: %v", err)
	}

	// Expect JSON format formatted with 2 spaces
	expected := `{
  "z": {
    "c": "nested_first",
    "a": "nested_second"
  },
  "a": "second"
}`
	if strings.TrimSpace(res) != strings.TrimSpace(expected) {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, res)
	}
}

func TestPrettifyYAMLLogic(t *testing.T) {
	yamlInput := `
z:
  c: nested_first
  a: nested_second
a: second
`
	res, err := PrettifyYAMLLogic(yamlInput, "2", "double")
	if err != nil {
		t.Fatalf("PrettifyYAMLLogic error: %v", err)
	}

	expected := `z:
  c: nested_first
  a: nested_second
a: second
`
	if strings.TrimSpace(res) != strings.TrimSpace(expected) {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, res)
	}
}

func TestValidateYAMLLogic(t *testing.T) {
	validYAML := "foo: bar\nlist:\n  - item"
	msg, valid, err := ValidateYAMLLogic(validYAML)
	if err != nil || !valid {
		t.Errorf("Expected valid YAML, got %v, valid: %v, msg: %v", err, valid, msg)
	}

	invalidYAML := "foo: bar\n  invalid_indent: true"
	msg, valid, err = ValidateYAMLLogic(invalidYAML)
	if err != nil || valid {
		t.Errorf("Expected invalid YAML, got %v, valid: %v, msg: %v", err, valid, msg)
	}
	if !strings.Contains(msg, "Invalid YAML") {
		t.Errorf("Expected message to contain 'Invalid YAML', got %v", msg)
	}
}
