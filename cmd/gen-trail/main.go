package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/pkg/domain"
)

func main() {
	targetDir := "examples/golden-path"
	if len(os.Args) > 1 {
		targetDir = os.Args[1]
	}

	// Ensure dir exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		panic(err)
	}

	fmt.Printf("Generating Golden Path in: %s\n", targetDir)

	// Init Loam (No Versioning = pure file generation)
	// This acts as our "Level Editor" saving to disk.
	repo, err := loam.Init(targetDir, loam.WithVersioning(false))
	if err != nil {
		panic(err)
	}

	typedRepo := loam.NewTyped[adapters.NodeMetadata](repo)
	ctx := context.TODO()

	// 1. Start (Clean)
	introMeta := adapters.NodeMetadata{
		ID:   "start",
		Type: "text",
		Transitions: []domain.Transition{
			{ToNodeID: "choice", Condition: ""},
		},
	}
	err = typedRepo.Save(ctx, &loam.DocumentModel[adapters.NodeMetadata]{
		ID:      "start",
		Content: "Welcome to the Golden Path.\nThis file is clean.",
		Data:    introMeta,
	})
	check(err)

	// 2. Choice (With Trailing Newlines/Noise)
	choiceMeta := adapters.NodeMetadata{
		ID:   "choice",
		Type: "question",
		Transitions: []domain.Transition{
			{ToNodeID: "end", Condition: "input == 'yes'"},
			{ToNodeID: "start", Condition: "input == 'no'"},
		},
	}
	// Injecting noise (trailing newlines and spaces) to verify Trellis/Loam trim logic
	err = typedRepo.Save(ctx, &loam.DocumentModel[adapters.NodeMetadata]{
		ID:      "choice",
		Content: "Do you want to finish?\n(Type 'yes' or 'no')\n\n\n   ",
		Data:    choiceMeta,
	})
	check(err)

	// 3. End
	endMeta := adapters.NodeMetadata{
		ID:          "end",
		Type:        "text",
		Transitions: []domain.Transition{},
	}
	err = typedRepo.Save(ctx, &loam.DocumentModel[adapters.NodeMetadata]{
		ID:      "end",
		Content: "The End.",
		Data:    endMeta,
	})
	check(err)

	fmt.Println("Done. Verify contents in", targetDir)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
