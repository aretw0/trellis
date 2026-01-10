package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/adapters/mcp"
	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run the Model Context Protocol (MCP) server",
	Long: `Starts the Trellis engine as an MCP Server.
This allows AI agents (like Claude Desktop) to interact with Trellis flows as tools.

Supported Transports:
- stdio (default): Uses Standard Input/Output. Ideal for local process integration.
- sse: Uses Server-Sent Events over HTTP. Ideal for remote agents or debuggers.`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("dir")
		if !cmd.Flags().Changed("dir") && len(args) > 0 {
			repoPath = args[0]
		}

		transport, _ := cmd.Flags().GetString("transport")
		port, _ := cmd.Flags().GetInt("port")

		// 1. Initialize Engine
		// Use ReadOnly mode implicitly via trellis.New (which sets Loam to ReadOnly)
		engine, err := trellis.New(repoPath)
		if err != nil {
			log.Fatalf("Error initializing trellis: %v", err)
		}

		// Configure logger
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
		slog.SetDefault(logger)

		// 2. Initialize MCP Server Adapter
		srv := mcp.NewServer(engine, engine.Loader())

		// 3. Start Server based on Transport
		switch transport {
		case "stdio":
			// Ensure logs don't corrupt JSON-RPC on Stdout
			log.SetOutput(os.Stderr)
			slog.Info("Starting Trellis MCP Server (Stdio)...")
			if err := srv.ServeStdio(); err != nil {
				slog.Error("MCP Server execution failed", "error", err)
				os.Exit(1)
			}
		case "sse":
			slog.Info("Starting Trellis MCP Server (SSE)", "port", port)

			// Create a context that cancels on interrupt signal
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if err := srv.ServeSSE(ctx, port); err != nil {
				// Ignore server closed error if it was caused by context cancellation
				if err != http.ErrServerClosed {
					slog.Error("MCP Server execution failed", "error", err)
					os.Exit(1)
				}
			}
			slog.Info("MCP Server stopped gracefully")
		default:
			log.Fatalf("Unknown transport: %s. Supported: stdio, sse", transport)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.Flags().String("transport", "stdio", "Transport protocol to use: 'stdio' or 'sse'")
	mcpCmd.Flags().Int("port", 8080, "Port to listen on (only for SSE)")
}
