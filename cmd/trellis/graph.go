package main

import (
	"fmt"
	"os"

	"context"
	"path/filepath"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/graph"
	"github.com/spf13/cobra"

	"github.com/aretw0/trellis/internal/adapters/file"
)

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Export the flow graph visualization",
	Long:  `Inspects the repository and outputs a Mermaid diagram (graph TD) representing the flow logic.`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("dir")
		if !cmd.Flags().Changed("dir") && len(args) > 0 {
			repoPath = args[0]
		}

		engine, err := trellis.New(repoPath)
		if err != nil {
			fmt.Printf("Error initializing trellis: %v\n", err)
			os.Exit(1)
		}

		nodes, err := engine.Inspect()
		if err != nil {
			fmt.Printf("Error inspecting graph: %v\n", err)
			os.Exit(1)
		}

		var overlay *graph.GraphOverlay
		sessionID, _ := cmd.Flags().GetString("session")
		if sessionID != "" {
			// Try to load session
			// Assumption: Sessions are stored in .trellis/sessions relative to CWD
			// This matches the default in examples and session cmd
			wd, _ := os.Getwd()
			storePath := filepath.Join(wd, ".trellis", "sessions")
			store := file.New(storePath)

			state, err := store.Load(context.Background(), sessionID)
			if err != nil {
				fmt.Printf("Error loading session '%s': %v\n", sessionID, err)
				os.Exit(1)
			}
			overlay = &graph.GraphOverlay{
				VisitedNodes: state.History,
				CurrentNode:  state.CurrentNodeID,
			}
		}

		// Generate and print Mermaid graph
		output := graph.GenerateMermaid(nodes, overlay)
		fmt.Print(output)
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().String("session", "", "Overlay session state (history & current node) on the graph")
}
