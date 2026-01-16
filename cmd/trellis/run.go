package main

import (
	"encoding/json"
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
		debug, _ := cmd.Flags().GetBool("debug")
		contextStr, _ := cmd.Flags().GetString("context")
		sessionID, _ := cmd.Flags().GetString("session")

		var initialContext map[string]any
		if contextStr != "" {
			if err := json.Unmarshal([]byte(contextStr), &initialContext); err != nil {
				fmt.Printf("Error parsing --context JSON: %v\n", err)
				os.Exit(1)
			}
		}

		if watchMode && headless {
			fmt.Println("Error: --watch and --headless cannot be used together.")
			os.Exit(1)
		}

		if watchMode {
			cli.RunWatch(repoPath, sessionID, debug)
		} else {
			err := cli.RunSession(repoPath, headless, jsonMode, debug, initialContext, sessionID)
			if err != nil {
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
	runCmd.Flags().Bool("debug", false, "Enable verbose debug logging (observability hooks)")
	runCmd.Flags().StringP("context", "c", "", "Initial context JSON string (e.g. '{\"user\": \"Alice\"}')")
	runCmd.Flags().StringP("session", "s", "", "Session ID for durable execution (resumes if exists)")
	runCmd.Flags().BoolP("watch", "w", false, "Run in development mode with hot-reload")

	// Make 'run' the default subcommand if no other command is provided.
	// This allows users to type 'trellis .' instead of 'trellis run .'
	rootCmd.Run = runCmd.Run
}
