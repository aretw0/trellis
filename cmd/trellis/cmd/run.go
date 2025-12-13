package cmd

import (
	"fmt"
	"os"

	"github.com/aretw0/trellis"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the interactive documentation flow",
	Long:  `Starts the Trellis engine in interactive mode with the content from the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("dir")

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

		// Execute
		if err := runner.Run(engine); err != nil {
			fmt.Printf("Error running trellis: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Make 'run' the default if no command is provided?
	// For now, let's keep it explicit, OR we can handle it in Run of rootCmd.
	// Common pattern: if no args, print help.
	// But Trellis legacy behavior was "run by default".
	// Let's add the logic to rootCmd.Run to default to runCmd behavior if desired.
	// However, for correct Cobra usage, we usually prefer explicit subcommands or a default Run action.
	// Let's configure rootCmd to run this logic if no subcommand given to preserve DX.

	rootCmd.Run = runCmd.Run
}
