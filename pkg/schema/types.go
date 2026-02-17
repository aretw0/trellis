package schema

import (
	"fmt"
	"reflect"
)

// Type defines the contract for field validation.
// Implementations determine how values are validated against a type.
type Type interface {
	// Name returns the human-readable name of the type (e.g., "string", "int").
	Name() string
	// Validate checks if a value conforms to this type.
	Validate(value any) error
}

// --- Built-in Type Implementations ---

// StringType validates string values.
type StringType struct{}

func (t *StringType) Name() string { return "string" }

func (t *StringType) Validate(value any) error {
	_, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	return nil
}

// IntType validates integer values.
type IntType struct{}

func (t *IntType) Name() string { return "int" }

func (t *IntType) Validate(value any) error {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return nil
	case float64:
		// Accept floats that are whole numbers (from JSON unmarshaling)
		if v == float64(int64(v)) {
			return nil
		}
		return fmt.Errorf("expected int, got float (not a whole number)")
	default:
		return fmt.Errorf("expected int, got %T", value)
	}
}

// FloatType validates floating-point values.
type FloatType struct{}

func (t *FloatType) Name() string { return "float" }

func (t *FloatType) Validate(value any) error {
	switch value.(type) {
	case float32, float64, int, int8, int16, int32, int64:
		return nil
	default:
		return fmt.Errorf("expected float, got %T", value)
	}
}

// BoolType validates boolean values.
type BoolType struct{}

func (t *BoolType) Name() string { return "bool" }

func (t *BoolType) Validate(value any) error {
	_, ok := value.(bool)
	if !ok {
		return fmt.Errorf("expected bool, got %T", value)
	}
	return nil
}

// SliceType validates slices of a specific element type.
type SliceType struct {
	elemType Type
}

func (t *SliceType) Name() string {
	return fmt.Sprintf("[%s]", t.elemType.Name())
}

func (t *SliceType) Validate(value any) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return fmt.Errorf("expected slice, got %T", value)
	}

	// Validate each element
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if err := t.elemType.Validate(elem); err != nil {
			return fmt.Errorf("element %d: %w", i, err)
		}
	}
	return nil
}

// CustomType applies a user-defined validation function.
type CustomType struct {
	name     string
	validate func(any) error
}

func (t *CustomType) Name() string { return t.name }

func (t *CustomType) Validate(value any) error {
	return t.validate(value)
}

// --- Factory Functions ---

// String creates a string type validator.
func String() Type { return &StringType{} }

// Int creates an integer type validator.
func Int() Type { return &IntType{} }

// Float creates a float type validator.
func Float() Type { return &FloatType{} }

// Bool creates a boolean type validator.
func Bool() Type { return &BoolType{} }

// Slice creates a slice type validator for elements of the given type.
func Slice(elemType Type) Type {
	return &SliceType{elemType: elemType}
}

// Custom creates a custom type validator with a user-defined function.
func Custom(name string, validate func(any) error) Type {
	return &CustomType{name: name, validate: validate}
}

// ParseType converts a string type name to a Type.
// Supports basic types: "string", "int", "float", "bool", "[string]", "[int]", etc.
func ParseType(typeStr string) (Type, error) {
	// Handle slice types: [string], [int], etc.
	if len(typeStr) > 2 && typeStr[0] == '[' && typeStr[len(typeStr)-1] == ']' {
		elemTypeStr := typeStr[1 : len(typeStr)-1]
		elemType, err := ParseType(elemTypeStr)
		if err != nil {
			return nil, err
		}
		return Slice(elemType), nil
	}

	// Handle built-in types
	switch typeStr {
	case "string":
		return String(), nil
	case "int":
		return Int(), nil
	case "float":
		return Float(), nil
	case "bool":
		return Bool(), nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", typeStr)
	}
}

// ParseTypeMap converts a map of field names to type strings into a Schema.
// Example: {"api_key": "string", "retries": "int"}
func ParseTypeMap(typeMap map[string]string) (Schema, error) {
	result := make(Schema)
	for key, typeStr := range typeMap {
		t, err := ParseType(typeStr)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", key, err)
		}
		result[key] = t
	}
	return result, nil
}
