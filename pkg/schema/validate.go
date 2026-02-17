package schema

// Schema is a map of field names to their expected types.
// Example: {"api_key": String(), "retries": Int(), "tags": Slice(String())}
type Schema map[string]Type

// Validate checks if data conforms to the schema.
// Returns an error with all validation failures found.
func Validate(schema Schema, data map[string]any) error {
	if len(schema) == 0 {
		// No schema = no validation
		return nil
	}

	var errs []error

	// Validate each field in the schema
	for fieldName, fieldType := range schema {
		value, exists := data[fieldName]
		if !exists {
			errs = append(errs, &ValidationError{
				Key:    fieldName,
				Reason: "required",
				Value:  nil,
			})
			continue
		}

		// Validate the value against the type
		if err := fieldType.Validate(value); err != nil {
			errs = append(errs, &ValidationError{
				Key:    fieldName,
				Reason: err.Error(),
				Value:  value,
			})
		}
	}

	// If there are errors, aggregate them
	if len(errs) > 0 {
		return &AggregateError{Errors: errs}
	}

	return nil
}

// ValidateFields validates only specific fields from data against the schema.
// Missing fields are treated as an error.
func ValidateFields(schema Schema, data map[string]any, fields ...string) error {
	if len(fields) == 0 {
		// No fields to validate
		return nil
	}

	var errs []error

	for _, fieldName := range fields {
		fieldType, exists := schema[fieldName]
		if !exists {
			// Field not defined in schema
			errs = append(errs, &ValidationError{
				Key:    fieldName,
				Reason: "not defined in schema",
				Value:  nil,
			})
			continue
		}

		value, fieldExists := data[fieldName]
		if !fieldExists {
			errs = append(errs, &ValidationError{
				Key:    fieldName,
				Reason: "required",
				Value:  nil,
			})
			continue
		}

		if err := fieldType.Validate(value); err != nil {
			errs = append(errs, &ValidationError{
				Key:    fieldName,
				Reason: err.Error(),
				Value:  value,
			})
		}
	}

	if len(errs) > 0 {
		return &AggregateError{Errors: errs}
	}

	return nil
}
