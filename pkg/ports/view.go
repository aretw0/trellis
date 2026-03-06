package ports

import "context"

// ContentConverter defines a generic interface for transforming rendered content.
// The engine applies this as a post-processing step after interpolation, remaining
// agnostic of what the conversion does (e.g., Markdown to HTML, sanitization, i18n).
type ContentConverter interface {
	// Convert takes raw content and returns the transformed representation.
	Convert(ctx context.Context, content string) (string, error)
}
