package main

import (
	"fmt"
	"os"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/compiler"
	"github.com/aretw0/trellis/internal/validator"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check the graph for consistency",
	Long:  `Crawls the graph starting from 'start' node and reports dead links or unreachable nodes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runValidate(args); err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Graph is valid! ✅")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(args []string) error {
	var dir string
	var err error

	if len(args) > 0 {
		dir = args[0]
	} else {
		dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// 1. Init Trellis Engine
	// We use the Engine to handle Loam initialization (which enforces ReadOnly by default).
	eng, err := trellis.New(dir)
	if err != nil {
		return fmt.Errorf("failed to init engine: %w", err)
	}

	// 2. Run Validation
	// We instantiate a parser to validate node content during traversal.
	parser := compiler.NewParser()

	if err := validator.ValidateGraph(eng.Loader(), parser, "start"); err != nil {
		return err
	}

	return nil
}
