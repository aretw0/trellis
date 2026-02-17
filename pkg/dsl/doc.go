/*
Package dsl provides a Go DSL (Domain Specific Language) for programmatically constructing Trellis graphs.

It allows developers to define complex state machine flows using a type-safe, fluent builder pattern
instead of relying on external YAML or JSON files. This is particularly useful for dynamic graph
generation, unit testing, and leveraging IDE autocompletion/type-checking.

Example usage:

	package main

	import (
		"github.com/aretw0/trellis/pkg/dsl"
	)

	func main() {
		pipeline := dsl.NewPipeline("my-flow")

		pipeline.Text("start").
			Content("Welcome to Trellis!").
			To("ask_name")

		pipeline.Input("ask_name").
			Content("What is your name?").
			SaveTo("user_name").
			To("end")

		pipeline.Text("end").
			Content("Goodbye, {{.user_name}}!").
			Terminal()

		// The resulting pipeline can be used as a ports.GraphLoader
		loader := pipeline.Build()
		// ... pass loader to trellis.New(...)
	}
*/
package dsl
