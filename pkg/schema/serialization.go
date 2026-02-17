package schema

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON serializes the schema as a map of field names to type strings.
func (s Schema) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	raw := make(map[string]string, len(s))
	for key, typ := range s {
		if typ == nil {
			return nil, fmt.Errorf("field %s: type is nil", key)
		}
		raw[key] = typ.Name()
	}

	return json.Marshal(raw)
}

// UnmarshalJSON deserializes the schema from a map of field names to type strings.
func (s *Schema) UnmarshalJSON(data []byte) error {
	if s == nil {
		return fmt.Errorf("schema: UnmarshalJSON on nil pointer")
	}

	if string(data) == "null" {
		*s = nil
		return nil
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		// Fallback: try map[string]any for cases where JSON decodes to mixed types
		var rawAny map[string]any
		if errAny := json.Unmarshal(data, &rawAny); errAny != nil {
			return err
		}
		raw = make(map[string]string, len(rawAny))
		for key, value := range rawAny {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("field %s: expected string type, got %T", key, value)
			}
			raw[key] = str
		}
	}

	parsed, err := ParseTypeMap(raw)
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}
