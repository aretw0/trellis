// Package schema provides a type-safe validation system for structured data.
//
// It defines a simple type system with built-in types (string, int, float, bool)
// and support for slices and custom validators. Schemas map field names to types,
// enabling runtime validation of complex data structures.
//
// Basic usage:
//
//	schema := schema.Schema{
//	    "api_key": schema.String(),
//	    "retries": schema.Int(),
//	    "tags":    schema.Slice(schema.String()),
//	}
//	
//	data := map[string]any{
//	    "api_key": "secret123",
//	    "retries": 3,
//	    "tags":    []string{"prod", "critical"},
//	}
//	
//	if err := schema.Validate(schema, data); err != nil {
//	    // Handle validation errors
//	}
//
// Schemas can be created programmatically or parsed from type strings:
//
//	typeMap := map[string]string{
//	    "api_key": "string",
//	    "retries": "int",
//	    "tags":    "[string]",
//	}
//	
//	schema, err := schema.ParseTypeMap(typeMap)
//
// Custom validators can be registered for domain-specific validation:
//
//	positiveInt := schema.Custom("positive_int", func(v any) error {
//	    i, ok := v.(int)
//	    if !ok {
//	        return fmt.Errorf("expected int")
//	    }
//	    if i <= 0 {
//	        return fmt.Errorf("must be positive")
//	    }
//	    return nil
//	})
//
// This package is designed to be library-agnostic, with zero external dependencies
// beyond the Go standard library. It can be embedded in larger systems or extracted
// as a standalone library.
package schema
