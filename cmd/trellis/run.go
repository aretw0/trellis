package main

import (
	"fmt"
	"os"

	"github.com/aretw0/trellis/internal/cli"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the interactive documentation flow",
	Long:  `Starts the Trellis engine in interactive mode with the content from the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("dir")
		if !cmd.Flags().Changed("dir") && len(args) > 0 {
			repoPath = args[0]
		}
		headless, _ := cmd.Flags().GetBool("headless")
		watchMode, _ := cmd.Flags().GetBool("watch")
		jsonMode, _ := cmd.Flags().GetBool("json")

		if watchMode && headless {
			fmt.Println("Error: --watch and --headless cannot be used together.")
			os.Exit(1)
		}

		if watchMode {
			cli.RunWatch(repoPath)
		} else {
			// cli.RunInteractive doesn't support passing handler yet.
			// We need to instantiate the runner manually or update RunInteractive.
			// Let's check RunInteractive in session.go.
			// Actually, let's just inline the runner setup here or update session.go?
			// Updating session.go is cleaner.
			if err := cli.RunSession(repoPath, headless, jsonMode); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().Bool("headless", false, "Run in headless mode (no prompts, strict IO)")
	runCmd.Flags().Bool("json", false, "Run in JSON mode (NDJSON input/output)")
	runCmd.Flags().BoolP("watch", "w", false, "Run in development mode with hot-reload")

	// Make 'run' the default if no command is provided?
	rootCmd.Run = runCmd.Run
}
