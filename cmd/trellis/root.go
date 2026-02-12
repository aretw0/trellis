package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "trellis",
	Short: "Trellis is a state machine based documentation engine",
	Long:  `Trellis allows you to build interactive documentation flows using simple Markdown files.`,
	Args:  cobra.ArbitraryArgs,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Persistent flags (available to all commands)
	rootCmd.PersistentFlags().String("dir", ".", "Directory containing the Trellis project")
	rootCmd.PersistentFlags().Bool("headless", false, "Run in headless mode (no prompts, strict IO)")
	rootCmd.PersistentFlags().Bool("json", false, "Run in JSON mode (NDJSON input/output)")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable verbose debug logging (observability hooks)")
	rootCmd.PersistentFlags().StringP("context", "c", "", "Initial context JSON string (e.g. '{\"user\": \"Alice\"}')")
	rootCmd.PersistentFlags().StringP("session", "s", "", "Session ID for durable execution (resumes if exists)")
	rootCmd.PersistentFlags().BoolP("watch", "w", false, "Run in development mode with hot-reload")
	rootCmd.PersistentFlags().Bool("fresh", false, "Start with a clean session (deletes existing session data)")
	rootCmd.PersistentFlags().String("redis-url", "", "Redis connection URL (e.g. redis://localhost:6379) for distributed state & locking")
	rootCmd.PersistentFlags().String("tools", "tools.yaml", "Path to the tool registry file")
	rootCmd.PersistentFlags().Bool("unsafe-inline", false, "Allow inline execution of scripts defined in Markdown (Dangerous)")
}
