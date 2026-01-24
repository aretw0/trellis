package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	// This script simulates a tool that reads environment variables
	// and outputs JSON.

	// Expect TRELLIS_ARG_NAME
	name := os.Getenv("TRELLIS_ARG_NAME")
	if name == "" {
		name = "World"
	}

	// Output formatted JSON
	output := map[string]string{
		"greeting": fmt.Sprintf("Hello, %s!", name),
		"source":   "Process Adapter",
	}

	// Pretty print to verification
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(output)
}
