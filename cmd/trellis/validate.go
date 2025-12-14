package main

import (
	"fmt"
	"os"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/validator"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check the graph for consistency",
	Long:  `Crawls the graph starting from 'start' node and reports dead links or unreachable nodes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runValidate(); err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Graph is valid! ✅")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 1. Init Loam
	repo, err := loam.Init(dir)
	if err != nil {
		return fmt.Errorf("failed to init loam: %w", err)
	}

	// 2. Run Validation
	if err := validator.ValidateGraph(repo, "start"); err != nil {
		return err
	}

	return nil
}
