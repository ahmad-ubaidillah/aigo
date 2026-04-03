package schema

import (
	"fmt"
	"strings"
)

func Validate(schema *Schema, input map[string]any) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if err := validateRequiredFields(schema, input); err != nil {
		return err
	}

	return validateAllProperties(schema, input)
}

func validateRequiredFields(schema *Schema, input map[string]any) error {
	for _, field := range schema.Required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("field %q: is required", field)
		}
	}

	return nil
}

func validateAllProperties(schema *Schema, input map[string]any) error {
	for name, propSchema := range schema.Properties {
		value, exists := input[name]
		if !exists {
			continue
		}

		if err := validateValue(propSchema, value); err != nil {
			return fmt.Errorf("field %q: %w", name, err)
		}
	}

	return nil
}

func validateValue(schema *Schema, value any) error {
	if schema == nil {
		return nil
	}

	validator := createValidator(schema)
	if validator == nil {
		return nil
	}

	return validator.Validate(value)
}

func createValidator(schema *Schema) TypeValidator {
	switch schema.Type {
	case "string":
		return &StringValidator{
			MinLength: schema.MinLength,
			MaxLength: schema.MaxLength,
			Pattern:   schema.Pattern,
			Enum:      schema.Enum,
		}
	case "number":
		return &NumberValidator{
			Minimum: schema.Minimum,
			Maximum: schema.Maximum,
			Enum:    schema.Enum,
		}
	case "integer":
		return &IntegerValidator{
			Minimum: schema.Minimum,
			Maximum: schema.Maximum,
			Enum:    schema.Enum,
		}
	case "boolean":
		return &BooleanValidator{}
	case "array":
		return &ArrayValidator{Items: schema.Items}
	case "object":
		return &ObjectValidator{
			Properties: schema.Properties,
			Required:   schema.Required,
		}
	default:
		return nil
	}
}

func ValidateWithPath(schema *Schema, input map[string]any, basePath string) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if err := validateRequiredFieldsPath(schema, input, basePath); err != nil {
		return err
	}

	return validatePropertiesPath(schema, input, basePath)
}

func validateRequiredFieldsPath(schema *Schema, input map[string]any, basePath string) error {
	for _, field := range schema.Required {
		if _, exists := input[field]; !exists {
			path := buildPath(basePath, field)
			return fmt.Errorf("field %q: is required", path)
		}
	}

	return nil
}

func validatePropertiesPath(schema *Schema, input map[string]any, basePath string) error {
	for name, propSchema := range schema.Properties {
		value, exists := input[name]
		if !exists {
			continue
		}

		path := buildPath(basePath, name)
		if err := validateValuePath(propSchema, value, path); err != nil {
			return err
		}
	}

	return nil
}

func validateValuePath(schema *Schema, value any, path string) error {
	if schema == nil {
		return nil
	}

	switch schema.Type {
	case "object":
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("field %q: expected object, got %T", path, value)
		}
		return ValidateWithPath(schema, obj, path)
	case "array":
		return validateArrayPath(schema, value, path)
	default:
		validator := createValidator(schema)
		if validator == nil {
			return nil
		}
		if err := validator.Validate(value); err != nil {
			return fmt.Errorf("field %q: %w", path, err)
		}
	}

	return nil
}

func validateArrayPath(schema *Schema, value any, basePath string) error {
	arr, ok := value.([]any)
	if !ok {
		return fmt.Errorf("field %q: expected array, got %T", basePath, value)
	}

	if schema.Items == nil {
		return nil
	}

	for i, item := range arr {
		path := fmt.Sprintf("%s[%d]", basePath, i)
		if err := validateValuePath(schema.Items, item, path); err != nil {
			return err
		}
	}

	return nil
}

func buildPath(basePath, field string) string {
	if basePath == "" {
		return field
	}

	if strings.HasPrefix(field, "[") {
		return basePath + field
	}

	return basePath + "." + field
}
