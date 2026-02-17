package schema

import (
	"fmt"
	"testing"
)

func TestStringType(t *testing.T) {
	typ := String()

	if typ.Name() != "string" {
		t.Errorf("Name() = %q, want %q", typ.Name(), "string")
	}

	tests := []struct {
		value   any
		wantErr bool
	}{
		{"hello", false},
		{"", false},
		{42, true},
		{3.14, true},
		{true, true},
		{nil, true},
	}

	for _, tt := range tests {
		err := typ.Validate(tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestIntType(t *testing.T) {
	typ := Int()

	if typ.Name() != "int" {
		t.Errorf("Name() = %q, want %q", typ.Name(), "int")
	}

	tests := []struct {
		value   any
		wantErr bool
	}{
		{42, false},
		{int8(42), false},
		{int16(42), false},
		{int32(42), false},
		{int64(42), false},
		{float64(42), false},     // whole number
		{float64(42.5), true},    // not whole
		{"42", true},
		{true, true},
		{nil, true},
	}

	for _, tt := range tests {
		err := typ.Validate(tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestFloatType(t *testing.T) {
	typ := Float()

	if typ.Name() != "float" {
		t.Errorf("Name() = %q, want %q", typ.Name(), "float")
	}

	tests := []struct {
		value   any
		wantErr bool
	}{
		{3.14, false},
		{float32(3.14), false},
		{42, false},
		{int64(42), false},
		{"3.14", true},
		{true, true},
		{nil, true},
	}

	for _, tt := range tests {
		err := typ.Validate(tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestBoolType(t *testing.T) {
	typ := Bool()

	if typ.Name() != "bool" {
		t.Errorf("Name() = %q, want %q", typ.Name(), "bool")
	}

	tests := []struct {
		value   any
		wantErr bool
	}{
		{true, false},
		{false, false},
		{1, true},
		{"true", true},
		{nil, true},
	}

	for _, tt := range tests {
		err := typ.Validate(tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestSliceType(t *testing.T) {
	stringSlice := Slice(String())
	intSlice := Slice(Int())
	stringStringSlice := Slice(Slice(String()))

	tests := []struct {
		typ     Type
		value   any
		wantErr bool
		desc    string
	}{
		// String slices
		{stringSlice, []string{"a", "b"}, false, "string slice"},
		{stringSlice, []string{}, false, "empty string slice"},
		{stringSlice, []interface{}{"a", "b"}, false, "any slice with strings"},
		{stringSlice, []int{1, 2}, true, "slice of ints when expecting strings"},
		{stringSlice, "not a slice", true, "string instead of slice"},
		// Int slices
		{intSlice, []int{1, 2, 3}, false, "int slice"},
		{intSlice, []interface{}{1, 2, 3}, false, "any slice with ints"},
		{intSlice, []interface{}{1, "2", 3}, true, "mixed slice"},
		// Nested slices
		{stringStringSlice, [][]string{{"a"}, {"b"}}, false, "nested string slice"},
		{stringStringSlice, [][]string{{"a"}, {"b", "c"}}, false, "nested string slice different lengths"},
	}

	for _, tt := range tests {
		err := tt.typ.Validate(tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: Validate(%v) error = %v, wantErr %v", tt.desc, tt.value, err, tt.wantErr)
		}
	}
}

func TestCustomType(t *testing.T) {
	evenNumber := Custom("even", func(v any) error {
		i, ok := v.(int)
		if !ok {
			return ErrCustomValidation("not an int")
		}
		if i%2 != 0 {
			return ErrCustomValidation("not even")
		}
		return nil
	})

	if evenNumber.Name() != "even" {
		t.Errorf("Name() = %q, want %q", evenNumber.Name(), "even")
	}

	tests := []struct {
		value   any
		wantErr bool
	}{
		{2, false},
		{4, false},
		{1, true},
		{3, true},
		{"2", true},
	}

	for _, tt := range tests {
		err := evenNumber.Validate(tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		wantName string
	}{
		{"string", false, "string"},
		{"int", false, "int"},
		{"float", false, "float"},
		{"bool", false, "bool"},
		{"[string]", false, "[string]"},
		{"[int]", false, "[int]"},
		{"[[string]]", false, "[[string]]"},
		{"invalid", true, ""},
		{"[invalid]", true, ""},
	}

	for _, tt := range tests {
		typ, err := ParseType(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && typ.Name() != tt.wantName {
			t.Errorf("ParseType(%q) Name() = %q, want %q", tt.input, typ.Name(), tt.wantName)
		}
	}
}

func TestParseTypeMap(t *testing.T) {
	typeMap := map[string]string{
		"api_key": "string",
		"retries": "int",
		"timeout": "float",
		"enabled": "bool",
		"tags": "[string]",
	}

	schema, err := ParseTypeMap(typeMap)
	if err != nil {
		t.Fatalf("ParseTypeMap() error = %v", err)
	}

	if len(schema) != len(typeMap) {
		t.Errorf("ParseTypeMap() len = %d, want %d", len(schema), len(typeMap))
	}

	if schema["api_key"].Name() != "string" {
		t.Error("api_key type should be string")
	}
	if schema["retries"].Name() != "int" {
		t.Error("retries type should be int")
	}
	if schema["tags"].Name() != "[string]" {
		t.Error("tags type should be [string]")
	}
}

func TestParseTypeMapError(t *testing.T) {
	typeMap := map[string]string{
		"api_key": "invalid",
	}

	_, err := ParseTypeMap(typeMap)
	if err == nil {
		t.Fatal("ParseTypeMap() should return error for invalid type")
	}
}

// Helper function for custom validators
func ErrCustomValidation(msg string) error {
	return fmt.Errorf("%s", msg)
}
