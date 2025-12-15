package main

import (
	"log"
	"os"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

func main() {
	// 1. Define the Graph using Go Structs
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    "question",
			Content: []byte("Welcome to the **Trellis TUI Demo**!\nDo you see this in *rich text*? [yes] [no]"),
			Transitions: []domain.Transition{
				{ToNodeID: "yes", Condition: "input == 'yes'"},
				{ToNodeID: "no"},
			},
		},
		domain.Node{
			ID:      "yes",
			Type:    "text",
			Content: []byte("## Great!\n\nThat means `glamour` is working."),
		},
		domain.Node{
			ID:      "no",
			Type:    "text",
			Content: []byte("## Oops.\n\nSomething is wrong with the renderer."),
		},
	)
	if err != nil {
		log.Fatalf("Failed to create loader: %v", err)
	}

	// 2. Initialize Engine
	eng, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		log.Fatalf("Failed to init engine: %v", err)
	}

	// 3. Configure Runner with TUI
	runner := trellis.NewRunner()
	runner.Input = os.Stdin
	runner.Output = os.Stdout
	runner.Headless = false
	runner.Renderer = tui.NewRenderer() // Inject TUI Renderer

	// 4. Run!
	if err := runner.Run(eng); err != nil {
		log.Fatalf("Error running: %v", err)
	}
}
