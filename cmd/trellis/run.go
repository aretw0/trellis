package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aretw0/lifecycle"
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
		fresh, _ := cmd.Flags().GetBool("fresh")
		redisURL, _ := cmd.Flags().GetString("redis-url")
		toolsPath, _ := cmd.Flags().GetString("tools")
		unsafeInline, _ := cmd.Flags().GetBool("unsafe-inline")

		opts := cli.RunOptions{
			RepoPath:     repoPath,
			Headless:     headless,
			Watch:        watchMode,
			JSON:         jsonMode,
			Debug:        debug,
			Context:      contextStr,
			SessionID:    sessionID,
			Fresh:        fresh,
			RedisURL:     redisURL,
			ToolsPath:    toolsPath,
			UnsafeInline: unsafeInline,
		}

		var lifecycleOpts []any
		// For all modes, we want the runner to handle SIGINT (Ctrl+C) to interrupt generation
		// without killing the application context immediately.
		lifecycleOpts = append(lifecycleOpts, lifecycle.WithCancelOnInterrupt(false))
		// Set ForceExit to 0: We give Trellis (via InteractiveRouter) full control over the exit strategy.
		lifecycleOpts = append(lifecycleOpts, lifecycle.WithForceExit(0))

		err := lifecycle.Run(lifecycle.Job(func(ctx context.Context) error {
			return cli.Execute(ctx, opts)
		}), lifecycleOpts...)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Make 'run' the default subcommand if no other command is provided.
	// This allows users to type 'trellis .' instead of 'trellis run .'
	rootCmd.Run = runCmd.Run
}
