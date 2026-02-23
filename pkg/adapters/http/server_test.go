package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aretw0/trellis/pkg/domain"
)

// MockEngine for testing
type MockEngine struct {
	WatchFunc func(ctx context.Context) (<-chan string, error)
	// Other methods are no-ops for this test
}

func (m *MockEngine) Render(ctx context.Context, state *domain.State) ([]domain.ActionRequest, bool, error) {
	return nil, false, nil
}
func (m *MockEngine) Navigate(ctx context.Context, state *domain.State, input any) (*domain.State, error) {
	// Simple mock: return a new state with changed context to trigger diff
	newState := state.Snapshot()
	if newState.Context == nil {
		newState.Context = make(map[string]any)
	}
	newState.Context["foo"] = "bar"
	newState.History = append(newState.History, "next")
	return newState, nil
}
func (m *MockEngine) Signal(ctx context.Context, state *domain.State, signal string) (*domain.State, error) {
	return nil, nil
}
func (m *MockEngine) Inspect() ([]domain.Node, error) { return nil, nil }
func (m *MockEngine) Watch(ctx context.Context) (<-chan string, error) {
	if m.WatchFunc != nil {
		return m.WatchFunc(ctx)
	}
	ch := make(chan string)
	close(ch)
	return ch, nil
}

func TestSubscribeEvents_Global(t *testing.T) {
	mockEng := &MockEngine{
		WatchFunc: func(ctx context.Context) (<-chan string, error) {
			ch := make(chan string, 1)
			ch <- "reload"
			close(ch)
			return ch, nil
		},
	}
	handler := NewHandler(mockEng)

	req := httptest.NewRequest("GET", "/events", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "event: ping") {
		t.Error("Expected ping event")
	}
	if !strings.Contains(body, "data: reload") {
		t.Error("Expected reload data")
	}
}

func TestSubscribeEvents_Session(t *testing.T) {
	mockEng := &MockEngine{}
	handler := NewHandler(mockEng)

	// 1. Subscribe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wSub := httptest.NewRecorder()
	reqSub := httptest.NewRequest("GET", "/events?session_id=sess-1", nil).WithContext(ctx)

	done := make(chan struct{})
	go func() {
		defer close(done)
		handler.ServeHTTP(wSub, reqSub)
	}()

	time.Sleep(100 * time.Millisecond) // Wait for subscription to register

	// 2. Trigger Navigate
	state := domain.State{
		SessionID:     "sess-1",
		CurrentNodeID: "start",
	}

	// Construct Input using generated helper
	inputVal := NavigateRequestInput0("go")
	inputContainer := &NavigateRequest_Input{}
	if err := inputContainer.FromNavigateRequestInput0(inputVal); err != nil {
		t.Fatalf("Failed to construct input: %v", err)
	}

	reqBody := NavigateJSONRequestBody{
		State: mapStateFromDomain(state),
		Input: inputContainer,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	reqNav := httptest.NewRequest("POST", "/navigate", bytes.NewReader(bodyBytes))
	wNav := httptest.NewRecorder()

	handler.ServeHTTP(wNav, reqNav)

	if wNav.Code != http.StatusOK {
		t.Fatalf("Navigate failed: %d %s", wNav.Code, wNav.Body.String())
	}

	// 3. Stop subscription to flush
	cancel()
	<-done // Wait for handler to finish writing to wSub

	output := wSub.Body.String()

	if !strings.Contains(output, "event: ping") {
		t.Error("Expected initial ping")
	}

	// Expect Context Diff: "foo":"bar"
	if !strings.Contains(output, `"foo":"bar"`) {
		t.Error("Expected context diff in SSE output")
	}
}

func TestServer_UI(t *testing.T) {
	mockEng := &MockEngine{}
	handler := NewHandler(mockEng)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Request the root UI path
	res, err := http.Get(ts.URL + "/ui/")
	if err != nil {
		t.Fatalf("Failed to GET /ui/: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected Content-Type text/html, got %q", contentType)
	}
}

func TestSubscribeEvents_Stress(t *testing.T) {
	mockEng := &MockEngine{}
	handler := NewHandler(mockEng)
	serverHandler, _ := handler.(http.HandlerFunc)

	// Create multiple connections to the same session
	sessionID := "stress-sess"
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Spin up 5 SSE listeners
	const numListeners = 5
	errCh := make(chan error, numListeners)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < numListeners; i++ {
		go func() {
			req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL+"/events?session_id="+sessionID, nil)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				errCh <- err
				return
			}
			defer res.Body.Close()
			// Just consume body to keep connection alive
			buf := make([]byte, 1024)
			for {
				_, err := res.Body.Read(buf)
				if err != nil {
					break
				}
			}
		}()
	}

	time.Sleep(100 * time.Millisecond) // Allow listeners to connect

	// Concurrently broadcast updates
	const numUpdates = 100
	done := make(chan struct{}, numUpdates)
	for i := 0; i < numUpdates; i++ {
		go func(i int) {
			inputVal := NavigateRequestInput0("stress")
			inputContainer := &NavigateRequest_Input{}
			inputContainer.FromNavigateRequestInput0(inputVal)

			reqBody := NavigateJSONRequestBody{
				State: mapStateFromDomain(domain.State{SessionID: sessionID, CurrentNodeID: "start"}),
				Input: inputContainer,
			}
			bodyBytes, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/navigate", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			// We can't access server direct from testing if it's wrapped in CORS, but we can serve http.
			if serverHandler != nil {
				serverHandler.ServeHTTP(w, req)
			} else {
				handler.ServeHTTP(w, req)
			}
			done <- struct{}{}
		}(i)
	}

	// Wait for all broadcasts to complete
	for i := 0; i < numUpdates; i++ {
		<-done
	}

	cancel() // Close listeners
}
