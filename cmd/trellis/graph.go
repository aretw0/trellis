package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aretw0/trellis"
	"github.com/spf13/cobra"
)

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Export the flow graph visualization",
	Long:  `Inspects the repository and outputs a Mermaid diagram (graph TD) representing the flow logic.`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("dir")

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

		fmt.Println("graph TD")
		for _, node := range nodes {
			// Sanitize ID for Mermaid
			safeID := sanitizeMermaidID(node.ID)

			// Node Label
			label := fmt.Sprintf("%s[\"%s\"]", safeID, node.ID)
			if node.Type == "start" || node.ID == "start" {
				label = fmt.Sprintf("%s((\"%s\"))", safeID, node.ID)
			}
			fmt.Printf("    %s\n", label)

			// Transitions
			for _, t := range node.Transitions {
				safeTo := sanitizeMermaidID(t.ToNodeID)
				arrow := "-->"
				if t.Condition != "" {
					arrow = fmt.Sprintf("-- \"%s\" -->", t.Condition)
				}
				fmt.Printf("    %s %s %s\n", safeID, arrow, safeTo)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
}

func sanitizeMermaidID(id string) string {
	return strings.ReplaceAll(strings.ReplaceAll(id, ".", "_"), "-", "_")
}
