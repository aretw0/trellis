package main

import (
	"fmt"
	"os"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the interactive documentation flow",
	Long:  `Starts the Trellis engine in interactive mode with the content from the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("dir")
		headless, _ := cmd.Flags().GetBool("headless")

		// Initialize Engine
		engine, err := trellis.New(repoPath)
		if err != nil {
			fmt.Printf("Error initializing trellis: %v\n", err)
			os.Exit(1)
		}

		// Configure Runner
		runner := trellis.NewRunner()
		runner.Input = os.Stdin
		runner.Output = os.Stdout
		runner.Headless = headless

		if !headless {
			runner.Renderer = tui.NewRenderer()
		}

		// Execute
		if err := runner.Run(engine); err != nil {
			fmt.Printf("Error running trellis: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().Bool("headless", false, "Run in headless mode (no prompts, strict IO)")

	// Make 'run' the default if no command is provided?
	rootCmd.Run = runCmd.Run
}
