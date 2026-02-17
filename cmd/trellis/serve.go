package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aretw0/lifecycle"
	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/logging"
	httpAdapter "github.com/aretw0/trellis/pkg/adapters/http"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the stateless HTTP server",
	Long:  `Starts the Trellis engine in stateless server mode, exposing a JSON API over HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := lifecycle.Run(lifecycle.Job(func(ctx context.Context) error {
			dir, _ := cmd.Flags().GetString("dir")
			port, _ := cmd.Flags().GetString("port")

			// Initialize Logger
			level := slog.LevelInfo
			debug, _ := cmd.Flags().GetBool("debug")
			if debug {
				level = slog.LevelDebug
			}
			logger := logging.New(level)
			slog.SetDefault(logger)

			engine, err := trellis.New(dir)
			if err != nil {
				return fmt.Errorf("error initializing trellis: %w", err)
			}

			handler := httpAdapter.NewHandler(engine)

			srv := &http.Server{
				Addr:    ":" + port,
				Handler: handler,
			}

			// Channel to listen for errors coming from the listener.
			serverErrors := make(chan error, 1)

			lifecycle.Go(ctx, func(ctx context.Context) error {
				logger.Info("Starting Trellis Server", "addr", srv.Addr, "dir", dir)
				serverErrors <- srv.ListenAndServe()
				return nil
			})

			// Blocking wait for shutdown or error.
			select {
			case err := <-serverErrors:
				// Error when starting HTTP server.
				return fmt.Errorf("server error: %w", err)

			case <-ctx.Done():
				logger.Info("Start shutdown", "reason", ctx.Err())

				// Give outstanding requests a deadline for completion.
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// Asking listener to shut down and shed load.
				if err := srv.Shutdown(shutdownCtx); err != nil {
					logger.Error("Graceful shutdown failed", "timeout", "5s", "err", err)
					if err := srv.Close(); err != nil {
						logger.Error("Error killing server", "err", err)
					}
				}
				logger.Info("Trellis Server stopped gracefully")
				return nil
			}
		}))

		if err != nil {
			slog.Error("Trellis Server exited with error", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
}
