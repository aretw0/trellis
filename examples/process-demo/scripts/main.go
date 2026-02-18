package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	// This script simulates a tool that reads TRELLIS_ARGS
	// and outputs JSON.

	// 2026-02-17: Tool Argument Evolution - use TRELLIS_ARGS
	rawArgs := os.Getenv("TRELLIS_ARGS")
	if rawArgs == "" {
		rawArgs = "{}"
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
		args = make(map[string]interface{})
	}

	name := "World"
	if n, ok := args["name"].(string); ok && n != "" {
		name = n
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
