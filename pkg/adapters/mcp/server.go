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

// RenderResponse aligns with the OpenAPI schema and provides a unified structure across adapters.
type RenderResponse struct {
	State    *domain.State          `json:"state,omitempty" jsonschema_description:"The current state of the engine"`
	Actions  []domain.ActionRequest `json:"actions" jsonschema_description:"List of available actions"`
	Terminal bool                   `json:"terminal" jsonschema_description:"Indicates if this is a terminal state"`
}

// Engine defines the interface required by the MCP server to interact with Trellis.
type Engine interface {
	ports.StatelessEngine
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
	// TOOL: render_state
	renderTool := mcp.NewTool("render_state",
		mcp.WithDescription("Render the view for a given valid state. If state is omitted, renders the start node."),
		mcp.WithString("node_id", mcp.Description("The ID of the node to render (optional if state is provided)")),
		mcp.WithString("history", mcp.Description("JSON array of node IDs visited (optional)")),
		mcp.WithString("context", mcp.Description("JSON object representing the current context (optional)")),
		mcp.WithOutputSchema[RenderResponse](),
	)
	s.mcpServer.AddTool(renderTool, mcp.NewStructuredToolHandler(s.handleRenderState))

	// TOOL: navigate
	navigateTool := mcp.NewTool("navigate",
		mcp.WithDescription("Navigate to the next state based on input."),
		mcp.WithString("node_id", mcp.Required(), mcp.Description("Current node ID")),
		mcp.WithString("input", mcp.Required(), mcp.Description("User input string")),
		mcp.WithString("history", mcp.Description("JSON array of visit history")),
		mcp.WithString("context", mcp.Description("JSON object of context")),
		mcp.WithOutputSchema[RenderResponse](),
	)
	s.mcpServer.AddTool(navigateTool, mcp.NewStructuredToolHandler(s.handleNavigate))

	// TOOL: send_signal
	signalTool := mcp.NewTool("send_signal",
		mcp.WithDescription("Send a global signal (e.g., interrupt, cancel) to the state machine."),
		mcp.WithString("signal", mcp.Required(), mcp.Description("Signal name")),
		mcp.WithString("node_id", mcp.Required(), mcp.Description("Current node ID")),
		mcp.WithString("history", mcp.Description("JSON array of visit history")),
		mcp.WithString("context", mcp.Description("JSON object of context")),
		mcp.WithOutputSchema[RenderResponse](),
	)
	s.mcpServer.AddTool(signalTool, mcp.NewStructuredToolHandler(s.handleSignal))

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

// Handler methods for structured tools

func (s *Server) handleRenderState(ctx context.Context, request mcp.CallToolRequest, args map[string]interface{}) (RenderResponse, error) {
	state := &domain.State{
		Context: make(map[string]interface{}),
		History: []string{},
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
		return RenderResponse{}, fmt.Errorf("render failed: %w", err)
	}

	return RenderResponse{
		State:    state,
		Actions:  actions,
		Terminal: terminal,
	}, nil
}

func (s *Server) handleNavigate(ctx context.Context, request mcp.CallToolRequest, args map[string]interface{}) (RenderResponse, error) {
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
		return RenderResponse{}, fmt.Errorf("input rejected: %w", err)
	}

	rich, err := runner.NavigateAndRender(ctx, s.engine, state, clean)
	if err != nil && rich == nil {
		return RenderResponse{}, fmt.Errorf("navigate failed: %w", err)
	}
	if err != nil {
		slog.Error("MCP Navigate: Render failed", "error", err)
	}

	return RenderResponse{
		State:    rich.State,
		Actions:  rich.Actions,
		Terminal: rich.Terminal,
	}, nil
}

func (s *Server) handleSignal(ctx context.Context, request mcp.CallToolRequest, args map[string]interface{}) (RenderResponse, error) {
	signal, _ := args["signal"].(string)
	nodeID, _ := args["node_id"].(string)

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
	}

	rich, err := runner.SignalAndRender(ctx, s.engine, state, signal)
	if err != nil && rich == nil {
		return RenderResponse{}, fmt.Errorf("signal failed: %w", err)
	}
	if err != nil {
		slog.Error("MCP Signal: Render failed", "error", err)
	}

	return RenderResponse{
		State:    rich.State,
		Actions:  rich.Actions,
		Terminal: rich.Terminal,
	}, nil
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
