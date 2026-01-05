package main

import (
	"fmt"
	"strings"

	"github.com/aretw0/trellis"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of trellis",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("trellis version %s\n", strings.TrimSpace(trellis.Version))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
