package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aretw0/trellis/internal/adapters/file"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage persistent sessions (Chaos Control)",
	Long:  `List, inspect, and remove persistent sessions stored in .trellis/sessions.`,
}

var sessionLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all active sessions",
	Run: func(cmd *cobra.Command, args []string) {
		store := getStore(cmd)
		sessions, err := store.List(cmd.Context())
		if err != nil {
			fmt.Printf("Error listing sessions: %v\n", err)
			os.Exit(1)
		}

		if len(sessions) == 0 {
			fmt.Println("No active sessions found.")
			return
		}

		fmt.Println("Active Sessions:")
		for _, s := range sessions {
			fmt.Println("- " + s)
		}
	},
}

var sessionInspectCmd = &cobra.Command{
	Use:   "inspect <session-id>",
	Short: "Inspect the state of a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		store := getStore(cmd)

		state, err := store.Load(cmd.Context(), sessionID)
		if err != nil {
			fmt.Printf("Error loading session '%s': %v\n", sessionID, err)
			os.Exit(1)
		}

		// Pretty print JSON
		data, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling state: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(data))
	},
}

var sessionRmCmd = &cobra.Command{
	Use:   "rm <session-id>...",
	Short: "Remove one or more sessions",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		store := getStore(cmd)
		hasError := false

		for _, sessionID := range args {
			if err := store.Delete(cmd.Context(), sessionID); err != nil {
				fmt.Printf("Error removing '%s': %v\n", sessionID, err)
				hasError = true
			} else {
				fmt.Printf("Removed session '%s'\n", sessionID)
			}
		}

		if hasError {
			os.Exit(1)
		}
	},
}

// TODO: Add support for --all flag in rm command

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(sessionLsCmd)
	sessionCmd.AddCommand(sessionInspectCmd)
	sessionCmd.AddCommand(sessionRmCmd)
}

func getStore(cmd *cobra.Command) *file.FileStore {
	projectDir, _ := cmd.Flags().GetString("dir")
	if projectDir == "" {
		projectDir = "."
	}
	// Replicate logic or rely on FileStore default?
	// FileStore constructs based on basePath.
	// We want to target <projectDir>/.trellis/sessions
	storePath := filepath.Join(projectDir, ".trellis", "sessions")
	return file.New(storePath)
}
