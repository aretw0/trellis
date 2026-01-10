package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

		srv := &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		}

		// Channel to listen for errors coming from the listener.
		serverErrors := make(chan error, 1)

		go func() {
			fmt.Printf("Starting Trellis Server on %s\n", srv.Addr)
			fmt.Printf("Serving content from: %s\n", dir)
			serverErrors <- srv.ListenAndServe()
		}()

		// Channel to listen for interrupt or terminate signals.
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		// Blocking main and waiting for shutdown.
		select {
		case err := <-serverErrors:
			// Error when starting HTTP server.
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)

		case sig := <-shutdown:
			fmt.Printf("\nStart shutdown... Signal: %v\n", sig)

			// Give outstanding requests a deadline for completion.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Asking listener to shut down and shed load.
			if err := srv.Shutdown(ctx); err != nil {
				fmt.Printf("Graceful shutdown did not complete in %v: %v\n", 5*time.Second, err)
				if err := srv.Close(); err != nil {
					fmt.Printf("Error killing server: %v\n", err)
				}
			}
			fmt.Println("Trellis Server stopped gracefully")
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
}
