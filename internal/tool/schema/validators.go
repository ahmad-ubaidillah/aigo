package schema

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

type TypeValidator interface {
	Validate(value any) error
}

type StringValidator struct {
	MinLength *int
	MaxLength *int
	Pattern   string
	Enum      []any
}

func (v *StringValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	if err := v.validateLength(str); err != nil {
		return err
	}

	if err := v.validatePattern(str); err != nil {
		return err
	}

	return v.validateEnum(str)
}

func (v *StringValidator) validateLength(str string) error {
	len := len(str)

	if v.MinLength != nil && len < *v.MinLength {
		return fmt.Errorf("string length %d is less than minimum %d", len, *v.MinLength)
	}

	if v.MaxLength != nil && len > *v.MaxLength {
		return fmt.Errorf("string length %d exceeds maximum %d", len, *v.MaxLength)
	}

	return nil
}

func (v *StringValidator) validatePattern(str string) error {
	if v.Pattern == "" {
		return nil
	}

	matched, err := regexp.MatchString(v.Pattern, str)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", v.Pattern, err)
	}

	if !matched {
		return fmt.Errorf("string %q does not match pattern %q", str, v.Pattern)
	}

	return nil
}

func (v *StringValidator) validateEnum(str string) error {
	if len(v.Enum) == 0 {
		return nil
	}

	for _, e := range v.Enum {
		if enumStr, ok := e.(string); ok && enumStr == str {
			return nil
		}
	}

	return fmt.Errorf("value %q is not one of allowed values: %v", str, v.Enum)
}

type NumberValidator struct {
	Minimum *float64
	Maximum *float64
	Enum    []any
}

func (v *NumberValidator) Validate(value any) error {
	num, ok := toNumeric(value)
	if !ok {
		return fmt.Errorf("expected number, got %T", value)
	}

	if err := v.validateRange(num); err != nil {
		return err
	}

	return v.validateEnum(num)
}

func (v *NumberValidator) validateRange(num float64) error {
	if v.Minimum != nil && num < *v.Minimum {
		return fmt.Errorf("value %v is less than minimum %v", num, *v.Minimum)
	}

	if v.Maximum != nil && num > *v.Maximum {
		return fmt.Errorf("value %v exceeds maximum %v", num, *v.Maximum)
	}

	return nil
}

func (v *NumberValidator) validateEnum(num float64) error {
	if len(v.Enum) == 0 {
		return nil
	}

	for _, e := range v.Enum {
		if enumNum, ok := toNumeric(e); ok && enumNum == num {
			return nil
		}
	}

	return fmt.Errorf("value %v is not one of allowed values: %v", num, v.Enum)
}

type IntegerValidator struct {
	Minimum *float64
	Maximum *float64
	Enum    []any
}

func (v *IntegerValidator) Validate(value any) error {
	num, ok := toNumeric(value)
	if !ok {
		return fmt.Errorf("expected integer, got %T", value)
	}

	if math.Trunc(num) != num {
		return fmt.Errorf("expected integer, got float %v", num)
	}

	if err := v.validateRange(num); err != nil {
		return err
	}

	return v.validateEnum(num)
}

func (v *IntegerValidator) validateRange(num float64) error {
	if v.Minimum != nil && num < *v.Minimum {
		return fmt.Errorf("value %v is less than minimum %v", num, *v.Minimum)
	}

	if v.Maximum != nil && num > *v.Maximum {
		return fmt.Errorf("value %v exceeds maximum %v", num, *v.Maximum)
	}

	return nil
}

func (v *IntegerValidator) validateEnum(num float64) error {
	if len(v.Enum) == 0 {
		return nil
	}

	for _, e := range v.Enum {
		if enumNum, ok := toNumeric(e); ok && enumNum == num {
			return nil
		}
	}

	return fmt.Errorf("value %v is not one of allowed values: %v", num, v.Enum)
}

type BooleanValidator struct{}

func (v *BooleanValidator) Validate(value any) error {
	if _, ok := value.(bool); ok {
		return nil
	}

	return fmt.Errorf("expected boolean, got %T", value)
}

type ArrayValidator struct {
	Items *Schema
}

func (v *ArrayValidator) Validate(value any) error {
	arr, ok := value.([]any)
	if !ok {
		return fmt.Errorf("expected array, got %T", value)
	}

	if v.Items == nil {
		return nil
	}

	for i, item := range arr {
		if err := validateValue(v.Items, item); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}

	return nil
}

type ObjectValidator struct {
	Properties map[string]*Schema
	Required   []string
}

func (v *ObjectValidator) Validate(value any) error {
	obj, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("expected object, got %T", value)
	}

	if err := v.validateRequired(obj); err != nil {
		return err
	}

	return v.validateProperties(obj)
}

func (v *ObjectValidator) validateRequired(obj map[string]any) error {
	for _, field := range v.Required {
		if _, exists := obj[field]; !exists {
			return fmt.Errorf("missing required field %q", field)
		}
	}

	return nil
}

func (v *ObjectValidator) validateProperties(obj map[string]any) error {
	for name, propSchema := range v.Properties {
		val, exists := obj[name]
		if !exists {
			continue
		}

		if err := validateValue(propSchema, val); err != nil {
			return fmt.Errorf("field %q: %w", name, err)
		}
	}

	return nil
}

func toNumeric(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func formatTypeError(expected string, got any) string {
	return fmt.Sprintf("expected %s, got %T", expected, got)
}

func joinEnumValues(enum []any) string {
	strs := make([]string, len(enum))
	for i, e := range enum {
		strs[i] = fmt.Sprintf("%v", e)
	}
	return strings.Join(strs, ", ")
}
