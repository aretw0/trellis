package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aretw0/lifecycle"
	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
	"github.com/aretw0/trellis/pkg/runner"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Engine defines the interface required by the MCP server to interact with Trellis.
type Engine interface {
	Render(ctx context.Context, state *domain.State) ([]domain.ActionRequest, bool, error)
	Navigate(ctx context.Context, state *domain.State, input any) (*domain.State, error)
	Inspect() ([]domain.Node, error)
}

// Server wraps the Trellis Engine and exposes it as an MCP Server.
type Server struct {
	engine    Engine
	loader    ports.GraphLoader
	mcpServer *server.MCPServer
}

// NewServer creates a new MCP Server instance.
func NewServer(engine Engine, loader ports.GraphLoader) *Server {
	s := &Server{
		engine:    engine,
		loader:    loader,
		mcpServer: server.NewMCPServer("trellis-mcp", strings.TrimSpace(trellis.Version)),
	}
	s.registerTools()
	s.registerResources()
	return s
}

// ServeStdio starts the server on Stdin/Stdout.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// ServeSSE starts the server on the given port using SSE.
func (s *Server) ServeSSE(ctx context.Context, port int) error {
	addr := fmt.Sprintf(":%d", port)
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// Start the SSE server
	sseServer := server.NewSSEServer(s.mcpServer, server.WithBaseURL(baseURL))

	mux := http.NewServeMux()
	mux.Handle("/sse", corsMiddleware(sseServer.SSEHandler()))
	mux.Handle("/message", corsMiddleware(sseServer.MessageHandler()))

	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	lifecycle.Go(ctx, func(ctx context.Context) error {
		slog.Info("MCP Server listening (SSE)", "address", addr)
		serverErrors <- httpServer.ListenAndServe()
		return nil
	})

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		// Create a timeout context for the graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fmt.Println("\nShutdown signal received, shutting down server...")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
		return nil
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("CORS Middleware", "method", r.Method, "path", r.URL.Path)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Baggage, Sentry-Trace")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) registerTools() {
	// TOOL: render
	s.mcpServer.AddTool(mcp.NewTool("render_state",
		mcp.WithDescription("Render the view for a given valid state. If state is omitted, renders the start node."),
		mcp.WithString("node_id", mcp.Description("The ID of the node to render (optional if state is provided)")),
		mcp.WithString("history", mcp.Description("JSON array of node IDs visited (optional)")),
		mcp.WithString("history", mcp.Description("JSON array of node IDs visited (optional)")),
		mcp.WithString("context", mcp.Description("JSON object representing the current context (optional)")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		state := &domain.State{
			Context: make(map[string]interface{}),
			History: []string{},
		}

		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			args = make(map[string]interface{})
		}

		nodeID, _ := args["node_id"].(string)

		if histStr, ok := args["history"].(string); ok {
			_ = json.Unmarshal([]byte(histStr), &state.History)
		}

		if ctxStr, ok := args["context"].(string); ok {
			_ = json.Unmarshal([]byte(ctxStr), &state.Context)
		} else if memStr, ok := args["memory"].(string); ok {
			// Backwards compatibility for 'memory'
			_ = json.Unmarshal([]byte(memStr), &state.Context)
		}

		if nodeID != "" {
			state.CurrentNodeID = nodeID
		} else if len(state.History) > 0 {
			state.CurrentNodeID = state.History[len(state.History)-1]
		} else {
			state.CurrentNodeID = "start"
		}

		actions, terminal, err := s.engine.Render(ctx, state)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("render failed: %v", err)), nil
		}

		result := map[string]interface{}{
			"actions":  actions,
			"terminal": terminal,
		}

		jsonBytes, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// TOOL: navigate
	s.mcpServer.AddTool(mcp.NewTool("navigate",
		mcp.WithDescription("Navigate to the next state based on input."),
		mcp.WithString("node_id", mcp.Required(), mcp.Description("Current node ID")),
		mcp.WithString("input", mcp.Required(), mcp.Description("User input string")),
		mcp.WithString("history", mcp.Description("JSON array of visit history")),
		mcp.WithString("context", mcp.Description("JSON object of context")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments format"), nil
		}

		nodeID, _ := args["node_id"].(string)
		input, _ := args["input"].(string)

		state := &domain.State{
			CurrentNodeID: nodeID,
			Context:       make(map[string]interface{}),
			History:       []string{},
		}

		if histStr, ok := args["history"].(string); ok {
			_ = json.Unmarshal([]byte(histStr), &state.History)
		}
		if ctxStr, ok := args["context"].(string); ok {
			_ = json.Unmarshal([]byte(ctxStr), &state.Context)
		} else if memStr, ok := args["memory"].(string); ok {
			_ = json.Unmarshal([]byte(memStr), &state.Context)
		}

		// Sanitize Input
		clean, err := runner.SanitizeInput(input)
		if err != nil {
			slog.Warn("MCP Navigate: Input rejected", "error", err, "size", len(input))
			return mcp.NewToolResultError(fmt.Sprintf("input rejected: %v", err)), nil
		}

		newState, err := s.engine.Navigate(ctx, state, clean)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("navigate failed: %v", err)), nil
		}

		jsonBytes, _ := json.Marshal(newState)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// TOOL: get_graph
	s.mcpServer.AddTool(mcp.NewTool("get_graph",
		mcp.WithDescription("Get the full graph definition for introspection."),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nodes, err := s.engine.Inspect()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("inspect failed: %v", err)), nil
		}
		jsonBytes, _ := json.Marshal(nodes)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}

func (s *Server) registerResources() {
	// EXPOSE: trellis://graph
	s.mcpServer.AddResource(mcp.NewResource("trellis://graph", "Current Graph Definition",
		mcp.WithMIMEType("application/json"),
	), func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		nodes, err := s.engine.Inspect()
		if err != nil {
			return nil, fmt.Errorf("failed to inspect graph: %w", err)
		}
		jsonBytes, _ := json.Marshal(nodes)

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "trellis://graph",
				MIMEType: "application/json",
				Text:     string(jsonBytes),
			},
		}, nil
	})
}
