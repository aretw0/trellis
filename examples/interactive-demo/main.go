package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

func main() {
	// Define nodes in memory using the Domain API directly.
	// This demonstrates how to use Trellis without filesystem dependencies.
	startNode := domain.Node{
		ID:           "start",
		Type:         domain.NodeTypeQuestion,
		InputType:    "choice",
		InputOptions: []string{"Yes", "No", "Maybe"},
		InputDefault: "Yes",
		Content:      []byte("Do you want to proceed?"),
		Transitions: []domain.Transition{
			{ToNodeID: "end", Condition: "input == 'Yes'"},
			{ToNodeID: "end", Condition: "input == 'No'"},
		},
	}
	endNode := domain.Node{
		ID:      "end",
		Type:    domain.NodeTypeText,
		Content: []byte("Done."),
	}

	// Create Memory Loader
	loader, err := memory.NewFromNodes(startNode, endNode)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Engine with Custom Loader
	eng, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	state := eng.Start()

	// 1. Render First State
	actions, _, err := eng.Render(ctx, state)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("--- Render Output ---")
	for _, act := range actions {
		if act.Type == domain.ActionRenderContent {
			fmt.Printf("[TEXT] %s\n", act.Payload)
		}
		if act.Type == domain.ActionRequestInput {
			req, ok := act.Payload.(domain.InputRequest)
			if !ok {
				fmt.Println("Error: Invalid Payload for InputRequest")
				continue
			}
			fmt.Printf("[INPUT REQUEST] Type=%s Options=%v Default=%s\n", req.Type, req.Options, req.Default)
		}
	}
}
