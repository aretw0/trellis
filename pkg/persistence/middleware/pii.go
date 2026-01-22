package middleware

import (
	"context"
	"regexp"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

type piiMiddleware struct {
	next     ports.StateStore
	patterns []*regexp.Regexp
}

// NewPIIMiddleware creates a middleware that masks values of keys matching the patterns.
func NewPIIMiddleware(patternStrings []string) Middleware {
	patterns := make([]*regexp.Regexp, len(patternStrings))
	for i, p := range patternStrings {
		patterns[i] = regexp.MustCompile(p)
	}
	return func(next ports.StateStore) ports.StateStore {
		return &piiMiddleware{next: next, patterns: patterns}
	}
}

func (m *piiMiddleware) Save(ctx context.Context, sessionID string, state *domain.State) error {
	// 1. Deep Clone to avoid side effects on the in-memory state used by the Engine.
	cloned := *state
	cloned.Context = deepCopyMap(state.Context)
	cloned.SystemContext = deepCopyMap(state.SystemContext)

	// 2. Mask PII
	maskMap(cloned.Context, m.patterns)

	return m.next.Save(ctx, sessionID, &cloned)
}

func (m *piiMiddleware) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	return m.next.Load(ctx, sessionID)
}

func (m *piiMiddleware) Delete(ctx context.Context, sessionID string) error {
	return m.next.Delete(ctx, sessionID)
}

func (m *piiMiddleware) List(ctx context.Context) ([]string, error) {
	return m.next.List(ctx)
}

// Helpers

func deepCopyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		// Handle nested maps
		if subMap, ok := v.(map[string]any); ok {
			out[k] = deepCopyMap(subMap)
		} else {
			out[k] = v // shallow copy of value
		}
	}
	return out
}

func maskMap(m map[string]any, patterns []*regexp.Regexp) {
	for k, v := range m {
		// Check key against patterns
		for _, p := range patterns {
			if p.MatchString(k) {
				m[k] = "***"
				break
			}
		}

		// Recurse if map
		if subMap, ok := v.(map[string]any); ok {
			maskMap(subMap, patterns)
		}
	}
}
