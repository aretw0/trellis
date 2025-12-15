package tui

import (
	"github.com/charmbracelet/glamour"
)

// NewRenderer returns a function that renders markdown using glamour.
// It uses a dark theme by default, but could be configurable.
func NewRenderer() func(string) (string, error) {
	// Initialize renderer with standard dark style
	// In the future, we can inject style preferences here.
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(), // Automatically detect light/dark background
	)

	return func(markdown string) (string, error) {
		return r.Render(markdown)
	}
}
