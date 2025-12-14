package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "trellis",
	Short: "Trellis is a state machine based documentation engine",
	Long:  `Trellis allows you to build interactive documentation flows using simple Markdown files.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Persistent flags (available to all commands)
	rootCmd.PersistentFlags().String("dir", ".", "Directory containing the Trellis project")
}
