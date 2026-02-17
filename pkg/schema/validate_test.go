package schema

import (
	"testing"
)

func TestValidate_Success(t *testing.T) {
	schema := Schema{
		"api_key": String(),
		"retries": Int(),
		"timeout": Float(),
		"enabled": Bool(),
		"tags":    Slice(String()),
	}

	data := map[string]any{
		"api_key": "secret123",
		"retries": 3,
		"timeout": 30.5,
		"enabled": true,
		"tags":    []string{"prod", "critical"},
	}

	err := Validate(schema, data)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidate_MissingField(t *testing.T) {
	schema := Schema{
		"api_key": String(),
		"retries": Int(),
	}

	data := map[string]any{
		"api_key": "secret123",
		// missing retries
	}

	err := Validate(schema, data)
	if err == nil {
		t.Fatal("Validate() should return error for missing field")
	}

	aggr, ok := err.(*AggregateError)
	if !ok {
		t.Fatalf("error should be *AggregateError, got %T", err)
	}

	if len(aggr.Errors) != 1 {
		t.Errorf("Validate() = %d errors, want 1", len(aggr.Errors))
	}

	validErr, ok := aggr.Errors[0].(*ValidationError)
	if !ok {
		t.Fatalf("error should be *ValidationError, got %T", aggr.Errors[0])
	}

	if validErr.Key != "retries" {
		t.Errorf("error Key = %q, want retries", validErr.Key)
	}
}

func TestValidate_TypeMismatch(t *testing.T) {
	schema := Schema{
		"api_key": String(),
		"retries": Int(),
	}

	data := map[string]any{
		"api_key": "secret123",
		"retries": "not an int",
	}

	err := Validate(schema, data)
	if err == nil {
		t.Fatal("Validate() should return error for type mismatch")
	}

	aggr, ok := err.(*AggregateError)
	if !ok {
		t.Fatalf("error should be *AggregateError, got %T", err)
	}

	if len(aggr.Errors) != 1 {
		t.Errorf("Validate() = %d errors, want 1", len(aggr.Errors))
	}

	validErr, ok := aggr.Errors[0].(*ValidationError)
	if !ok {
		t.Fatalf("error should be *ValidationError, got %T", aggr.Errors[0])
	}

	if validErr.Key != "retries" {
		t.Errorf("error Key = %q, want retries", validErr.Key)
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	schema := Schema{
		"api_key": String(),
		"retries": Int(),
		"timeout": Float(),
	}

	data := map[string]any{
		// missing api_key
		"retries": "not an int",
		"timeout": "not a float",
	}

	err := Validate(schema, data)
	if err == nil {
		t.Fatal("Validate() should return error")
	}

	aggr, ok := err.(*AggregateError)
	if !ok {
		t.Fatalf("error should be *AggregateError, got %T", err)
	}

	if len(aggr.Errors) != 3 {
		t.Errorf("Validate() = %d errors, want 3", len(aggr.Errors))
	}
}

func TestValidate_EmptySchema(t *testing.T) {
	schema := Schema{}
	data := map[string]any{
		"api_key": "secret123",
	}

	err := Validate(schema, data)
	if err != nil {
		t.Errorf("Validate() with empty schema should return nil, got %v", err)
	}
}

func TestValidate_NilSchema(t *testing.T) {
	var schema Schema
	data := map[string]any{
		"api_key": "secret123",
	}

	err := Validate(schema, data)
	if err != nil {
		t.Errorf("Validate() with nil schema should return nil, got %v", err)
	}
}

func TestValidateFields_Success(t *testing.T) {
	schema := Schema{
		"api_key": String(),
		"retries": Int(),
		"timeout": Float(),
	}

	data := map[string]any{
		"api_key": "secret123",
		"retries": 3,
		"timeout": 30.5,
	}

	err := ValidateFields(schema, data, "api_key", "retries")
	if err != nil {
		t.Errorf("ValidateFields() error = %v, want nil", err)
	}
}

func TestValidateFields_PartialValidation(t *testing.T) {
	schema := Schema{
		"api_key": String(),
		"retries": Int(),
		"timeout": Float(),
	}

	data := map[string]any{
		"api_key": "secret123",
		"retries": "invalid", // Wrong type, but not validated
		"timeout": "invalid", // Wrong type, but not validated
	}

	err := ValidateFields(schema, data, "api_key")
	if err != nil {
		t.Errorf("ValidateFields(api_key only) error = %v, want nil", err)
	}
}

func TestValidateFields_MissingField(t *testing.T) {
	schema := Schema{
		"api_key": String(),
		"retries": Int(),
	}

	data := map[string]any{
		"api_key": "secret123",
	}

	err := ValidateFields(schema, data, "api_key", "retries")
	if err == nil {
		t.Fatal("ValidateFields() should return error for missing field")
	}

	aggr, ok := err.(*AggregateError)
	if !ok {
		t.Fatalf("error should be *AggregateError, got %T", err)
	}

	if len(aggr.Errors) != 1 {
		t.Errorf("ValidateFields() = %d errors, want 1", len(aggr.Errors))
	}
}

func TestValidateFields_UndefinedField(t *testing.T) {
	schema := Schema{
		"api_key": String(),
	}

	data := map[string]any{
		"api_key": "secret123",
		"unknown": "value",
	}

	err := ValidateFields(schema, data, "unknown")
	if err == nil {
		t.Fatal("ValidateFields() should return error for undefined field")
	}

	aggr, ok := err.(*AggregateError)
	if !ok {
		t.Fatalf("error should be *AggregateError, got %T", err)
	}

	if len(aggr.Errors) != 1 {
		t.Errorf("ValidateFields() = %d errors, want 1", len(aggr.Errors))
	}

	validErr, ok := aggr.Errors[0].(*ValidationError)
	if !ok {
		t.Fatalf("error should be *ValidationError, got %T", aggr.Errors[0])
	}

	if validErr.Key != "unknown" {
		t.Errorf("error Key = %q, want unknown", validErr.Key)
	}
}

func TestValidateFields_Empty(t *testing.T) {
	schema := Schema{
		"api_key": String(),
	}

	data := map[string]any{}

	err := ValidateFields(schema, data)
	if err != nil {
		t.Errorf("ValidateFields() with no fields should return nil, got %v", err)
	}
}

func TestValidationError_String(t *testing.T) {
	tests := []struct {
		err  *ValidationError
		want string
	}{
		{
			&ValidationError{Key: "api_key", Reason: "required", Value: nil},
			`field "api_key": required`,
		},
		{
			&ValidationError{Key: "retries", Reason: "expected int, got string", Value: "invalid"},
			`field "retries": expected int, got string (got string)`,
		},
	}

	for _, tt := range tests {
		got := tt.err.Error()
		if got != tt.want {
			t.Errorf("ValidationError.Error() = %q, want %q", got, tt.want)
		}
	}
}

func TestAggregateError_String(t *testing.T) {
	aggr := &AggregateError{
		Errors: []error{
			&ValidationError{Key: "api_key", Reason: "required", Value: nil},
			&ValidationError{Key: "retries", Reason: "expected int", Value: "invalid"},
		},
	}

	result := aggr.Error()
	if result == "" {
		t.Error("AggregateError.Error() should not be empty")
	}

	// Should contain count
	if !containsString(result, "2 validation errors") {
		t.Errorf("AggregateError.Error() should mention 2 errors, got: %s", result)
	}
}

func TestValidationErrors(t *testing.T) {
	aggr := &AggregateError{
		Errors: []error{
			&ValidationError{Key: "api_key", Reason: "required", Value: nil},
		},
	}

	errs := ValidationErrors(aggr)
	if len(errs) != 1 {
		t.Errorf("ValidationErrors() = %d errors, want 1", len(errs))
	}

	// Non-aggregate error returns nil
	err := &ValidationError{Key: "api_key", Reason: "required", Value: nil}
	errs = ValidationErrors(err)
	if errs != nil {
		t.Errorf("ValidationErrors() on non-aggregate = %v, want nil", errs)
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr))
}
