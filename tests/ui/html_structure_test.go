package ui_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpAdapter "github.com/aretw0/trellis/pkg/adapters/http"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// simpleMockEngine is a minimal implementation for UI testing
type simpleMockEngine struct {
	ports.StatelessEngine
}

func (m *simpleMockEngine) Render(ctx context.Context, state *domain.State) ([]domain.ActionRequest, bool, error) {
	return nil, false, nil
}
func (m *simpleMockEngine) Inspect() ([]domain.Node, error) { return nil, nil }
func (m *simpleMockEngine) Watch(ctx context.Context) (<-chan string, error) {
	return make(chan string), nil
}

func TestInspectorUI_HTMLStructure(t *testing.T) {
	handler := httpAdapter.NewHandler(&simpleMockEngine{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/ui/")
	if err != nil {
		t.Fatalf("Failed to fetch UI: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}
	html := string(bodyBytes)

	// Define required tags/attributes for A11y and I18n
	requiredPatterns := []struct {
		Pattern string
		Name    string
	}{
		{`id="lang-picker"`, "Language Picker"},
		{`class="skip-link"`, "Skip Link"},
		{`role="banner"`, "Header Landmark"},
		{`role="main"`, "Main Landmark"},
		{`role="log"`, "Chat Log Landmark"},
		{`aria-live="polite"`, "Aria Live polite"},
		{`data-i18n="title"`, "I18n Title Tag"},
		{`bubble-user`, "User bubble class"},
	}

	for _, req := range requiredPatterns {
		if !strings.Contains(html, req.Pattern) {
			t.Errorf("Missing required HTML pattern: %s (pattern: %s)", req.Name, req.Pattern)
		}
	}
}
