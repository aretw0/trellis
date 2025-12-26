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
		headless, _ := cmd.Flags().GetBool("headless")
		watchMode, _ := cmd.Flags().GetBool("watch")

		if watchMode && headless {
			fmt.Println("Error: --watch and --headless cannot be used together.")
			os.Exit(1)
		}

		if watchMode {
			cli.RunWatch(repoPath)
		} else {
			if err := cli.RunInteractive(repoPath, headless); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().Bool("headless", false, "Run in headless mode (no prompts, strict IO)")
	runCmd.Flags().BoolP("watch", "w", false, "Run in development mode with hot-reload")

	// Make 'run' the default if no command is provided?
	rootCmd.Run = runCmd.Run
}
