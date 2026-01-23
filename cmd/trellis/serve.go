package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		dir, _ := cmd.Flags().GetString("dir")
		port, _ := cmd.Flags().GetString("port")

		// Initialize Logger
		logger := logging.New(slog.LevelInfo)
		slog.SetDefault(logger)

		engine, err := trellis.New(dir)
		if err != nil {
			logger.Error("Error initializing trellis", "err", err)
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
			logger.Info("Starting Trellis Server", "addr", srv.Addr, "dir", dir)
			serverErrors <- srv.ListenAndServe()
		}()

		// Channel to listen for interrupt or terminate signals.
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		// Blocking main and waiting for shutdown.
		select {
		case err := <-serverErrors:
			// Error when starting HTTP server.
			logger.Error("Server error", "err", err)
			os.Exit(1)

		case sig := <-shutdown:
			logger.Info("Start shutdown", "signal", sig)

			// Give outstanding requests a deadline for completion.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Asking listener to shut down and shed load.
			if err := srv.Shutdown(ctx); err != nil {
				logger.Error("Graceful shutdown failed", "timeout", "5s", "err", err)
				if err := srv.Close(); err != nil {
					logger.Error("Error killing server", "err", err)
				}
			}
			logger.Info("Trellis Server stopped gracefully")
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
}
