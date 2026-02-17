package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/dsl"
	"github.com/aretw0/trellis/pkg/runner"
)

func main() {
	// 1. Define the graph using the Go DSL
	b := dsl.New()

	b.Add("start").
		Text("Hello! This is a graph built with Go.").
		Go("ask_name")

	b.Add("ask_name").
		Question("What is your name?").
		Input("text").
		SaveTo("name").
		Go("greet")

	b.Add("greet").
		Text("Nice to meet you, {{.name}}!").
		Go("end")

	b.Add("end").
		Text("Goodbye!")

	// 2. Compile to a Loader (in-memory)
	loader, err := b.Build()
	if err != nil {
		fmt.Printf("Error building graph: %v\n", err)
		os.Exit(1)
	}

	// 3. Initialize the Engine with the Loader
	engine, err := trellis.New("dsl-example",
		trellis.WithLoader(loader),
	)
	if err != nil {
		fmt.Printf("Error initializing engine: %v\n", err)
		os.Exit(1)
	}

	// 4. Run the graph using the Runner
	ctx := context.Background()
	handler := runner.NewTextHandler(os.Stdout, runner.WithStdin())
	r := runner.NewRunner(
		runner.WithEngine(engine),
		runner.WithInputHandler(handler),
	)

	if err := r.Run(ctx); err != nil {
		fmt.Printf("Error running graph: %v\n", err)
		os.Exit(1)
	}
}
