// Package schema provides JSON Schema-like validation for tool inputs.
package schema

import (
	"fmt"
)

// Schema represents a validation schema for tool input parameters.
type Schema struct {
	Type        string             `json:"type"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	MinLength   *int               `json:"minLength,omitempty"`
	MaxLength   *int               `json:"maxLength,omitempty"`
	Minimum     *float64           `json:"minimum,omitempty"`
	Maximum     *float64           `json:"maximum,omitempty"`
	Pattern     string             `json:"pattern,omitempty"`
	Enum        []any              `json:"enum,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Default     any                `json:"default,omitempty"`
	Description string             `json:"description,omitempty"`
}

// SchemaParser parses raw map data into Schema structures.
type SchemaParser struct{}

// NewSchemaParser creates a new SchemaParser instance.
func NewSchemaParser() *SchemaParser {
	return &SchemaParser{}
}

// Parse converts a raw map[string]any into a Schema structure.
func (p *SchemaParser) Parse(raw map[string]any) (*Schema, error) {
	if raw == nil {
		return nil, fmt.Errorf("schema cannot be nil")
	}

	schema := &Schema{Properties: make(map[string]*Schema)}

	if err := p.parseBasicFields(raw, schema); err != nil {
		return nil, err
	}
	if err := p.parseStringConstraints(raw, schema); err != nil {
		return nil, err
	}
	if err := p.parseNumberConstraints(raw, schema); err != nil {
		return nil, err
	}
	if err := p.parseProperties(raw, schema); err != nil {
		return nil, err
	}

	p.parseSimpleFields(raw, schema)
	return schema, nil
}

func (p *SchemaParser) parseBasicFields(raw map[string]any, schema *Schema) error {
	if typeVal, ok := raw["type"]; ok {
		typeStr, ok := typeVal.(string)
		if !ok {
			return fmt.Errorf("type must be a string, got %T", typeVal)
		}
		schema.Type = typeStr
	}
	return nil
}

func (p *SchemaParser) parseStringConstraints(raw map[string]any, schema *Schema) error {
	if minLen, ok := raw["minLength"]; ok {
		val, err := toInt(minLen)
		if err != nil {
			return fmt.Errorf("minLength: %w", err)
		}
		schema.MinLength = &val
	}

	if maxLen, ok := raw["maxLength"]; ok {
		val, err := toInt(maxLen)
		if err != nil {
			return fmt.Errorf("maxLength: %w", err)
		}
		schema.MaxLength = &val
	}

	if pattern, ok := raw["pattern"]; ok {
		patternStr, ok := pattern.(string)
		if !ok {
			return fmt.Errorf("pattern must be a string, got %T", pattern)
		}
		schema.Pattern = patternStr
	}

	return nil
}

func (p *SchemaParser) parseNumberConstraints(raw map[string]any, schema *Schema) error {
	if min, ok := raw["minimum"]; ok {
		val, err := toFloat64(min)
		if err != nil {
			return fmt.Errorf("minimum: %w", err)
		}
		schema.Minimum = &val
	}

	if max, ok := raw["maximum"]; ok {
		val, err := toFloat64(max)
		if err != nil {
			return fmt.Errorf("maximum: %w", err)
		}
		schema.Maximum = &val
	}

	return nil
}

func (p *SchemaParser) parseProperties(raw map[string]any, schema *Schema) error {
	if props, ok := raw["properties"]; ok {
		propsMap, ok := props.(map[string]any)
		if !ok {
			return fmt.Errorf("properties must be an object, got %T", props)
		}

		for name, propRaw := range propsMap {
			propMap, ok := propRaw.(map[string]any)
			if !ok {
				return fmt.Errorf("property %q must be an object, got %T", name, propRaw)
			}

			propSchema, err := p.Parse(propMap)
			if err != nil {
				return fmt.Errorf("property %q: %w", name, err)
			}
			schema.Properties[name] = propSchema
		}
	}

	if req, ok := raw["required"]; ok {
		reqArr, ok := req.([]any)
		if !ok {
			return fmt.Errorf("required must be an array, got %T", req)
		}

		schema.Required = make([]string, 0, len(reqArr))
		for i, r := range reqArr {
			rStr, ok := r.(string)
			if !ok {
				return fmt.Errorf("required[%d] must be a string, got %T", i, r)
			}
			schema.Required = append(schema.Required, rStr)
		}
	}

	if items, ok := raw["items"]; ok {
		itemsMap, ok := items.(map[string]any)
		if !ok {
			return fmt.Errorf("items must be an object, got %T", items)
		}

		itemsSchema, err := p.Parse(itemsMap)
		if err != nil {
			return fmt.Errorf("items: %w", err)
		}
		schema.Items = itemsSchema
	}

	return nil
}

func (p *SchemaParser) parseSimpleFields(raw map[string]any, schema *Schema) {
	if enum, ok := raw["enum"]; ok {
		if enumArr, ok := enum.([]any); ok {
			schema.Enum = enumArr
		}
	}

	if def, ok := raw["default"]; ok {
		schema.Default = def
	}

	if desc, ok := raw["description"]; ok {
		if descStr, ok := desc.(string); ok {
			schema.Description = descStr
		}
	}
}

func toInt(val any) (int, error) {
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		if v != float64(int(v)) {
			return 0, fmt.Errorf("value must be an integer")
		}
		return int(v), nil
	default:
		return 0, fmt.Errorf("value must be a number, got %T", val)
	}
}

func toFloat64(val any) (float64, error) {
	switch v := val.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("value must be a number, got %T", val)
	}
}
