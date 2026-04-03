package schema

import (
	"testing"
)

func ptrInt(i int) *int           { return &i }
func ptrFloat(f float64) *float64 { return &f }

func TestSchemaParser_Parse(t *testing.T) {
	t.Parallel()

	p := &SchemaParser{}
	raw := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":      "string",
				"minLength": float64(1),
				"maxLength": float64(100),
			},
			"age": map[string]any{
				"type":    "integer",
				"minimum": float64(0),
				"maximum": float64(150),
			},
			"tags": map[string]any{
				"type": "array",
			},
		},
		"required": []any{"name"},
		"enum":     []any{"a", "b"},
	}

	s, err := p.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.Type != "object" {
		t.Errorf("expected object, got %s", s.Type)
	}
	if len(s.Properties) != 3 {
		t.Errorf("expected 3 properties, got %d", len(s.Properties))
	}
	if len(s.Required) != 1 || s.Required[0] != "name" {
		t.Errorf("expected required=[name], got %v", s.Required)
	}
	if len(s.Enum) != 2 {
		t.Errorf("expected 2 enum values, got %d", len(s.Enum))
	}
}

func TestSchemaParser_ParseNested(t *testing.T) {
	t.Parallel()

	p := &SchemaParser{}
	raw := map[string]any{
		"properties": map[string]any{
			"nested": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"inner": map[string]any{"type": "string"},
				},
			},
		},
	}

	s, err := p.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.Properties["nested"] == nil {
		t.Error("expected nested property")
	}
}

func TestSchemaParser_ParseInvalidProperty(t *testing.T) {
	t.Parallel()

	p := &SchemaParser{}
	raw := map[string]any{
		"properties": map[string]any{
			"bad": "not a map",
		},
	}

	_, err := p.Parse(raw)
	if err == nil {
		t.Error("expected error for invalid property")
	}
}

func TestValidate_ObjectValid(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name": {Type: "string", MinLength: ptrInt(1)},
			"age":  {Type: "integer", Minimum: ptrFloat(0)},
		},
		Required: []string{"name"},
	}

	err := Validate(s, map[string]any{"name": "test", "age": float64(25)})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_MissingRequired(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type:     "object",
		Required: []string{"name"},
	}
	err := Validate(s, map[string]any{})
	if err == nil {
		t.Error("expected error for missing required")
	}
}

func TestValidate_StringMinLength(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name": {Type: "string", MinLength: ptrInt(5)},
		},
	}
	err := Validate(s, map[string]any{"name": "hi"})
	if err == nil {
		t.Error("expected error for short string")
	}
}

func TestValidate_StringMaxLength(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name": {Type: "string", MaxLength: ptrInt(3)},
		},
	}
	err := Validate(s, map[string]any{"name": "hello"})
	if err == nil {
		t.Error("expected error for long string")
	}
}

func TestValidate_NumberMinimum(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"val": {Type: "number", Minimum: ptrFloat(10)},
		},
	}
	err := Validate(s, map[string]any{"val": float64(5)})
	if err == nil {
		t.Error("expected error for below minimum")
	}
}

func TestValidate_NumberMaximum(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"val": {Type: "number", Maximum: ptrFloat(10)},
		},
	}
	err := Validate(s, map[string]any{"val": float64(15)})
	if err == nil {
		t.Error("expected error for above maximum")
	}
}

func TestValidate_Enum(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"choice": {Type: "string", Enum: []any{"a", "b"}},
		},
	}
	err := Validate(s, map[string]any{"choice": "c"})
	if err == nil {
		t.Error("expected error for value not in enum")
	}
	err = Validate(s, map[string]any{"choice": "a"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_Pattern(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"code": {Type: "string", Pattern: "^[a-z]+$"},
		},
	}
	err := Validate(s, map[string]any{"code": "hello"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = Validate(s, map[string]any{"code": "HELLO"})
	if err == nil {
		t.Error("expected error for pattern mismatch")
	}
}

func TestValidate_InvalidPattern(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"code": {Type: "string", Pattern: "[invalid"},
		},
	}
	err := Validate(s, map[string]any{"code": "test"})
	if err == nil {
		t.Error("expected error for invalid pattern")
	}
}

func TestValidate_WrongType(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name": {Type: "string"},
		},
	}
	err := Validate(s, map[string]any{"name": 123})
	if err == nil {
		t.Error("expected error for wrong type")
	}
}

func TestValidate_NilSchema(t *testing.T) {
	t.Parallel()

	err := Validate(nil, map[string]any{"a": "b"})
	if err == nil {
		t.Error("expected error for nil schema")
	}
}

func TestValidate_NilInput(t *testing.T) {
	t.Parallel()

	s := &Schema{Type: "object"}
	err := Validate(s, nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}

func TestValidate_ArrayType(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"items": {Type: "array"},
		},
	}
	err := Validate(s, map[string]any{"items": []any{"a", "b"}})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_BooleanType(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"flag": {Type: "boolean"},
		},
	}
	err := Validate(s, map[string]any{"flag": true})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = Validate(s, map[string]any{"flag": "not a bool"})
	if err == nil {
		t.Error("expected error for non-boolean")
	}
}

func TestValidate_UnknownType(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"val": {Type: "unknown"},
		},
	}
	err := Validate(s, map[string]any{"val": "test"})
	if err != nil {
		t.Errorf("expected no error for unknown type, got %v", err)
	}
}

func TestValidateWithPath_Valid(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name": {Type: "string", MinLength: ptrInt(1)},
		},
		Required: []string{"name"},
	}
	err := ValidateWithPath(s, map[string]any{"name": "test"}, "root")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateWithPath_MissingRequired(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type:     "object",
		Required: []string{"name"},
	}
	err := ValidateWithPath(s, map[string]any{}, "root")
	if err == nil {
		t.Error("expected error for missing required")
	}
}

func TestValidateWithPath_NilSchema(t *testing.T) {
	t.Parallel()

	err := ValidateWithPath(nil, map[string]any{"a": "b"}, "root")
	if err == nil {
		t.Error("expected error for nil schema")
	}
}

func TestValidateWithPath_NilInput(t *testing.T) {
	t.Parallel()

	s := &Schema{Type: "object"}
	err := ValidateWithPath(s, nil, "root")
	if err == nil {
		t.Error("expected error for nil input")
	}
}

func TestValidateWithPath_NestedObject(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"nested": {
				Type: "object",
				Properties: map[string]*Schema{
					"inner": {Type: "string"},
				},
			},
		},
	}
	err := ValidateWithPath(s, map[string]any{"nested": map[string]any{"inner": "hello"}}, "root")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateWithPath_ArrayValidation(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"items": {
				Type:  "array",
				Items: &Schema{Type: "string"},
			},
		},
	}
	err := ValidateWithPath(s, map[string]any{"items": []any{"a", "b"}}, "root")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateWithPath_ArrayInvalidType(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"items": {
				Type:  "array",
				Items: &Schema{Type: "string"},
			},
		},
	}
	err := ValidateWithPath(s, map[string]any{"items": "not an array"}, "root")
	if err == nil {
		t.Error("expected error for non-array")
	}
}

func TestValidateWithPath_ArrayItemValidation(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"items": {
				Type:  "array",
				Items: &Schema{Type: "string"},
			},
		},
	}
	err := ValidateWithPath(s, map[string]any{"items": []any{123}}, "root")
	if err == nil {
		t.Error("expected error for invalid array item")
	}
}

func TestBuildPath(t *testing.T) {
	t.Parallel()

	if buildPath("", "field") != "field" {
		t.Errorf("expected 'field', got %s", buildPath("", "field"))
	}
	if buildPath("root", "field") != "root.field" {
		t.Errorf("expected 'root.field', got %s", buildPath("root", "field"))
	}
	if buildPath("root", "[0]") != "root[0]" {
		t.Errorf("expected 'root[0]', got %s", buildPath("root", "[0]"))
	}
}
