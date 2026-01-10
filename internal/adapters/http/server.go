package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/go-chi/chi/v5"
)

//go:generate go tool oapi-codegen -package http -generate types,chi-server,spec -o api.gen.go ../../../api/openapi.yaml

// Engine defines the interface for the Trellis state machine core.
type Engine interface {
	Render(ctx context.Context, state *domain.State) ([]domain.ActionRequest, bool, error)
	Navigate(ctx context.Context, state *domain.State, input string) (*domain.State, error)
	Inspect() ([]domain.Node, error)
	Watch(ctx context.Context) (<-chan string, error)
}

// Server implements the generated ServerInterface
type Server struct {
	Engine Engine
}

// Ensure Server implements ServerInterface
var _ ServerInterface = (*Server)(nil)

// NewHandler creates a new HTTP handler for the engine.
func NewHandler(engine Engine) http.Handler {
	server := &Server{Engine: engine}
	r := chi.NewRouter()

	// Swagger UI
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		// Use the generated rawSpec function to get the embedded spec
		spec, err := rawSpec()
		if err != nil {
			http.Error(w, "Failed to load spec", http.StatusInternalServerError)
			return
		}
		w.Write(spec)
	})
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(swaggerHTML))
	})

	return HandlerFromMux(server, r)
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
		return
	}

	domainState := mapStateToDomain(body)
	actions, terminal, err := s.Engine.Render(r.Context(), &domainState)
	if err != nil {
		http.Error(w, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}

	resp := RenderResponse{
		Actions:  ptr(mapActionsFromDomain(actions)),
		Terminal: &terminal,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Printf("Render encode error: %v\n", err)
	}
}

// Navigate handles the POST /navigate request.
func (s *Server) Navigate(w http.ResponseWriter, r *http.Request) {
	var body NavigateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	domainState := mapStateToDomain(body.State)
	input := ""
	if body.Input != nil {
		input = *body.Input
	}

	newState, err := s.Engine.Navigate(r.Context(), &domainState, input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Navigate error: %v", err), http.StatusInternalServerError)
		return
	}

	resp := mapStateFromDomain(*newState)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Printf("Navigate encode error: %v\n", err)
	}
}

// GetGraph handles the GET /graph request.
func (s *Server) GetGraph(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.Engine.Inspect()
	if err != nil {
		http.Error(w, fmt.Sprintf("Inspect error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		fmt.Printf("GetGraph encode error: %v\n", err)
	}
}

// SubscribeEvents handles the GET /events request (SSE).
func (s *Server) SubscribeEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	events, err := s.Engine.Watch(r.Context())
	if err != nil {
		// Log error but we can't write to W anymore if we started flushing?
		// Actually we haven't flushed yet, so we can try.
		http.Error(w, fmt.Sprintf("Watch error: %v", err), http.StatusInternalServerError)
		return
	}

	// Send initial connection message? Optional. Let's send a ping.
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
			fmt.Printf("Visualizer Event: %s\n", event)
			fmt.Fprintf(w, "data: %s\n\n", event)
			flusher.Flush()
		}
	}
}

// -- Helpers --

func ptr[T any](v T) *T {
	return &v
}

func mapStateToDomain(s State) domain.State {
	d := domain.State{
		CurrentNodeID: s.CurrentNodeId,
		Memory:        make(map[string]any),
		History:       []string{},
		Terminated:    false,
	}
	if s.Memory != nil {
		d.Memory = *s.Memory
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
		CurrentNodeId: d.CurrentNodeID,
		Memory:        ptr(d.Memory),
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
