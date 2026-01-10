package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aretw0/trellis"
	httpAdapter "github.com/aretw0/trellis/internal/adapters/http"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the stateless HTTP server",
	Long:  `Starts the Trellis engine in stateless server mode, exposing a JSON API over HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		dir, _ := cmd.Flags().GetString("dir")
		port, _ := cmd.Flags().GetString("port")

		engine, err := trellis.New(dir)
		if err != nil {
			fmt.Printf("Error initializing trellis: %v\n", err)
			os.Exit(1)
		}

		handler := httpAdapter.NewHandler(engine)

		fmt.Printf("Starting Trellis Server on :%s\n", port)
		fmt.Printf("Serving content from: %s\n", dir)

		if err := http.ListenAndServe(":"+port, handler); err != nil {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
}
