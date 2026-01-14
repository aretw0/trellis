package main

import (
	"fmt"
	"os"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/graph"
	"github.com/spf13/cobra"
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

		// Generate and print Mermaid graph
		output := graph.GenerateMermaid(nodes)
		fmt.Print(output)
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
}
