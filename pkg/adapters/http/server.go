package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
	"github.com/go-chi/chi/v5"
)

//go:generate go tool oapi-codegen -package http -generate types,chi-server,spec -o api.gen.go ../../../api/openapi.yaml

// Engine defines the interface for the Trellis state machine core.
type Engine interface {
	Render(ctx context.Context, state *domain.State) ([]domain.ActionRequest, bool, error)
	Navigate(ctx context.Context, state *domain.State, input any) (*domain.State, error)
	Signal(ctx context.Context, state *domain.State, signal string) (*domain.State, error)
	Inspect() ([]domain.Node, error)
	Watch(ctx context.Context) (<-chan string, error)
}

// Server implements the generated ServerInterface
type Server struct {
	Engine  Engine
	Streams *StreamManager
}

// Ensure Server implements ServerInterface
var _ ServerInterface = (*Server)(nil)

// NewHandler creates a new HTTP handler for the engine.
func NewHandler(engine Engine) http.Handler {
	server := &Server{
		Engine:  engine,
		Streams: NewStreamManager(),
	}
	r := chi.NewRouter()

	// Swagger UI
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		// Use the generated rawSpec function to get the embedded spec
		spec, err := rawSpec()
		if err != nil {
			http.Error(w, "Failed to load spec", http.StatusInternalServerError)
			slog.Error("Failed to load OpenAPI spec", "error", err)
			return
		}
		w.Write(spec)
	})
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(swaggerHTML))
	})

	handler := HandlerFromMux(server, r)
	return enableCORS(handler)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Custom-Header")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

const swaggerHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Trellis API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
<script>
    window.onload = () => {
    window.ui = SwaggerUIBundle({
        url: '/openapi.yaml',
        dom_id: '#swagger-ui',
    });
    };
</script>
</body>
</html>
`

// Render handles the POST /render request.
func (s *Server) Render(w http.ResponseWriter, r *http.Request) {
	var body RenderJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		slog.Warn("Render: Invalid request body", "error", err)
		return
	}

	domainState := mapStateToDomain(body)
	actions, terminal, err := s.Engine.Render(r.Context(), &domainState)
	if err != nil {
		http.Error(w, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		slog.Error("Render failed", "error", err)
		return
	}

	resp := RenderResponse{
		Actions:  ptr(mapActionsFromDomain(actions)),
		Terminal: &terminal,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Render response encode failed", "error", err)
	}
}

// Navigate handles the POST /navigate request.
func (s *Server) Navigate(w http.ResponseWriter, r *http.Request) {
	var body NavigateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		slog.Warn("Navigate: Invalid request body", "error", err)
		return
	}

	domainState := mapStateToDomain(body.State)
	input := ""
	if body.Input != nil {
		var err error
		input, err = body.Input.AsNavigateRequestInput0()
		if err != nil {
			http.Error(w, "Invalid input format: expected string", http.StatusBadRequest)
			slog.Warn("Navigate: Invalid input format", "error", err)
			return
		}
	}

	// Sanitize Input (Global Policy)
	if input != "" {
		clean, err := runner.SanitizeInput(input)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid input: %v", err), http.StatusBadRequest)
			slog.Warn("Navigate: Input rejected", "error", err, "size", len(input))
			return
		}
		input = clean
	}

	newState, err := s.Engine.Navigate(r.Context(), &domainState, input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Navigate error: %v", err), http.StatusInternalServerError)
		slog.Error("Navigate failed", "error", err)
		return
	}

	// Calculate and Broadcast Diff
	diff := domain.Diff(&domainState, newState)
	if diff != nil {
		slog.Debug("Navigate: Diff calculated", "diff", diff, "session_id", domainState.SessionID)
		if bytes, err := json.Marshal(diff); err == nil {
			s.Streams.Broadcast(domainState.SessionID, string(bytes))
		}
	} else {
		slog.Debug("Navigate: No diff calculated", "session_id", domainState.SessionID)
	}

	resp := mapStateFromDomain(*newState)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Navigate response encode failed", "error", err)
	}
}

// Signal handles the POST /signal request.
func (s *Server) Signal(w http.ResponseWriter, r *http.Request) {
	var body SignalJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		slog.Warn("Signal: Invalid request body", "error", err)
		return
	}

	domainState := mapStateToDomain(body.State)

	newState, err := s.Engine.Signal(r.Context(), &domainState, body.Signal)
	if err != nil {
		if err == domain.ErrUnhandledSignal {
			http.Error(w, fmt.Sprintf("Signal unhandled: %v", err), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Signal error: %v", err), http.StatusInternalServerError)
		slog.Error("Signal failed", "error", err)
		return
	}

	// Calculate and Broadcast Diff
	diff := domain.Diff(&domainState, newState)
	if diff != nil {
		slog.Debug("Signal: Diff calculated", "diff", diff, "session_id", domainState.SessionID)
		if bytes, err := json.Marshal(diff); err == nil {
			s.Streams.Broadcast(domainState.SessionID, string(bytes))
		}
	} else {
		slog.Debug("Signal: No diff calculated", "session_id", domainState.SessionID)
	}

	resp := mapStateFromDomain(*newState)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Signal response encode failed", "error", err)
	}
}

// GetGraph handles the GET /graph request.
func (s *Server) GetGraph(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.Engine.Inspect()
	if err != nil {
		http.Error(w, fmt.Sprintf("Inspect error: %v", err), http.StatusInternalServerError)
		slog.Error("Inspect failed", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		slog.Error("GetGraph response encode failed", "error", err)
	}
}

// GetHealth handles the GET /health request.
func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetInfo handles the GET /info request.
func (s *Server) GetInfo(w http.ResponseWriter, r *http.Request) {
	apiVersion := "unknown"
	if swagger, err := GetSwagger(); err == nil && swagger.Info != nil {
		apiVersion = swagger.Info.Version
	}

	resp := map[string]string{
		"app":         "trellis-http",
		"version":     strings.TrimSpace(trellis.Version),
		"api_version": apiVersion,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// StreamManager handles active SSE connections
type StreamManager struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan<- string]struct{} // SessionID -> Set of Channels
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		subscribers: make(map[string]map[chan<- string]struct{}),
	}
}

func (sm *StreamManager) Subscribe(sessionID string) (chan string, func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ch := make(chan string, 10)
	if _, ok := sm.subscribers[sessionID]; !ok {
		sm.subscribers[sessionID] = make(map[chan<- string]struct{})
	}
	sm.subscribers[sessionID][ch] = struct{}{}

	return ch, func() {
		sm.mu.Lock()
		defer sm.mu.Unlock()
		if subs, ok := sm.subscribers[sessionID]; ok {
			delete(subs, ch)
			close(ch)
			if len(subs) == 0 {
				delete(sm.subscribers, sessionID)
			}
		}
	}
}

func (sm *StreamManager) Broadcast(sessionID string, msg string) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	slog.Debug("StreamManager: Broadcasting", "session_id", sessionID, "payload_size", len(msg))

	if subs, ok := sm.subscribers[sessionID]; ok {
		slog.Debug("StreamManager: Found subscribers", "count", len(subs))
		for ch := range subs {
			select {
			case ch <- msg:
			default:
				// Drop message if channel is full (slow client)
				slog.Warn("SSE: Client buffer full, dropping message", "session_id", sessionID)
			}
		}
	}
}

// SubscribeEvents handles the GET /events request (SSE).
func (s *Server) SubscribeEvents(w http.ResponseWriter, r *http.Request, params SubscribeEventsParams) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		slog.Error("SubscribeEvents: Streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Global Hot Reload (Legacy / No Session)
	if params.SessionId == nil {
		slog.Info("SSE: Subscribing to Global Hot Reload")
		events, err := s.Engine.Watch(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("Watch error: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", event)
				flusher.Flush()
			}
		}
	}

	// Session-based Subscription (State Diff)
	sessionID := *params.SessionId
	slog.Info("SSE: Subscribing to Session Updates", "session_id", sessionID)

	// StreamManager is initialized in NewHandler and attached to Server.

	ch, cancel := s.Streams.Subscribe(sessionID)
	defer cancel()

	fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
	flusher.Flush()

	// Parse 'watch' filter
	var watchList []string
	if params.Watch != nil {
		watchList = strings.Split(*params.Watch, ",")
	}

	for {
		select {
		case <-r.Context().Done():
			slog.Info("SSE Client Disconnected")
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			// Apply Filtering if provided.
			// Note: We currently deserialize JSON to check fields, which has a performance cost.
			// Future optimization: Push filtering down to the Broadcast level or send raw bytes if matching.
			if len(watchList) > 0 {
				var diff domain.StateDiff
				if err := json.Unmarshal([]byte(msg), &diff); err == nil {
					// Check if any watched field is present
					keep := false
					for _, field := range watchList {
						field = strings.TrimSpace(field)
						switch field {
						case "context":
							if len(diff.Context) > 0 {
								keep = true
							}
						case "history":
							if diff.HistoryParams != nil {
								keep = true
							}
						case "status":
							if diff.Status != nil || diff.Terminated != nil {
								keep = true
							}
						}
					}
					if !keep {
						continue
					}
				}
			}

			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func init() {
	// Configure default slog to output JSON to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
}

// -- Helpers --

func ptr[T any](v T) *T {
	return &v
}

func mapStateToDomain(s State) domain.State {
	d := domain.State{
		CurrentNodeID: s.CurrentNodeId,
		Context:       make(map[string]any),
		History:       []string{},
		Terminated:    false,
	}
	if s.SessionId != nil {
		d.SessionID = *s.SessionId
	}
	if s.Memory != nil {
		d.Context = *s.Memory
	}
	if s.History != nil {
		d.History = *s.History
	}
	if s.Terminated != nil {
		d.Terminated = *s.Terminated
	}
	return d
}

func mapStateFromDomain(d domain.State) State {
	s := State{
		SessionId:     ptr(d.SessionID),
		CurrentNodeId: d.CurrentNodeID,
		Memory:        ptr(d.Context),
		Terminated:    &d.Terminated,
	}
	if d.History != nil {
		s.History = &d.History
	}
	return s
}

func mapActionsFromDomain(actions []domain.ActionRequest) []ActionRequest {
	res := make([]ActionRequest, len(actions))
	for i, a := range actions {
		res[i] = ActionRequest{
			Type:    a.Type,
			Payload: a.Payload,
		}
	}
	return res
}
