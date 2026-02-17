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
		b := dsl.New()

		b.Add("start").
			Text("Welcome to Trellis!").
			Go("ask_name")

		b.Add("ask_name").
			Question("What is your name?").
			SaveTo("user_name").
			Go("end")

		b.Add("end").
			Text("Goodbye, {{.user_name}}!").
			Terminal()

		// The resulting builder can be compiled into a loader
		loader, err := b.Build()
		// ... pass loader to trellis.New(...)
	}
*/
package dsl
