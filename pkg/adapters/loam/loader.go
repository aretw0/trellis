package loam

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/mitchellh/mapstructure"
)

// Loader adapts the Loam library to the Trellis GraphLoader interface.
type Loader struct {
	Repo *loam.TypedRepository[NodeMetadata]
}

// New creates a new Loam adapter.
func New(repo *loam.TypedRepository[NodeMetadata]) *Loader {
	return &Loader{
		Repo: repo,
	}
}

// GetNode retrieves a node from the Loam repository using the direct Service API.
// Note: Loam Service.GetDocument is a direct convenience lookup.
func (l *Loader) GetNode(id string) ([]byte, error) {
	ctx := context.Background()

	// Loam Normalized Retrieval.
	// We trust Loam to find the file (e.g. start.md) even if we ask for "start",
	// or we assume the seeding created "start" (which maps to start.md).
	doc, err := l.Repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("loam get failed for %s: %w", id, err)
	}

	// Synthesize valid JSON for the compiler.
	// We must map our loader struct (which matches YAML "to") to domain JSON ("to_node_id").

	// Merge Transitions and Options
	totalTransitions := len(doc.Data.Transitions) + len(doc.Data.Options)
	domainTransitions := make([]domain.Transition, 0, totalTransitions)

	// Helper to convert LoaderTransition (DTO) to domain.Transition
	convert := func(lt LoaderTransition) domain.Transition {
		// Support both "to" and "to_node_id"
		to := lt.To
		if to == "" {
			to = lt.ToFull
		}
		if to == "" {
			to = lt.JumpTo
		}
		from := lt.From
		if from == "" {
			from = lt.FromFull
		}

		condition := lt.Condition
		// Sugar: Map "text" to condition if condition is empty
		// This supports the "options" syntax
		if condition == "" && lt.Text != "" {
			// Simple exact match logic aligned with DefaultEvaluator
			condition = fmt.Sprintf("input == '%s'", strings.ReplaceAll(lt.Text, "'", "\\'"))
		}

		return domain.Transition{
			FromNodeID: from,
			ToNodeID:   to,
			Condition:  condition,
		}
	}

	for _, opt := range doc.Data.Options {
		domainTransitions = append(domainTransitions, convert(opt))
	}
	for _, lt := range doc.Data.Transitions {
		domainTransitions = append(domainTransitions, convert(lt))
	}

	data := make(map[string]any)
	// Normalize ID: prefer metadata ID, fallback to filename ID, but always strip extension
	rawID := doc.Data.ID
	if rawID == "" {
		rawID = doc.ID
	}
	data["id"] = trimExtension(rawID)

	data["type"] = doc.Data.Type
	if doc.Data.Type == "" {
		data["type"] = domain.NodeTypeText
	}

	if doc.Data.Wait {
		data["wait"] = doc.Data.Wait
	}
	if doc.Data.OnError != "" {
		data["on_error"] = doc.Data.OnError
	}
	if len(doc.Data.OnSignal) > 0 {
		data["on_signal"] = doc.Data.OnSignal
	}
	if doc.Data.SaveTo != "" {
		data["save_to"] = doc.Data.SaveTo
	}
	if len(doc.Data.RequiredContext) > 0 {
		data["required_context"] = doc.Data.RequiredContext
	}
	if len(doc.Data.DefaultContext) > 0 {
		data["default_context"] = doc.Data.DefaultContext
	}
	data["transitions"] = domainTransitions
	data["content"] = []byte(doc.Content) // As Base64

	// Map Input Configuration
	if doc.Data.InputType != "" {
		data["input_type"] = doc.Data.InputType
		data["input_options"] = doc.Data.InputOptions
		data["input_default"] = doc.Data.InputDefault
	}

	// Map Tool Configuration
	// Priority: Do > ToolCall (Alias)
	var toolCall *domain.ToolCall
	if doc.Data.Do != nil {
		toolCall = doc.Data.Do
	} else if doc.Data.ToolCall != nil {
		toolCall = doc.Data.ToolCall
	}

	if toolCall != nil {
		// Ensure ID is present
		if toolCall.ID == "" {
			toolCall.ID = toolCall.Name
		}
		data["do"] = toolCall
	}
	if len(doc.Data.Tools) > 0 {
		resolvedTools, err := l.resolveTools(ctx, doc.Data.Tools, nil)
		if err != nil {
			return nil, fmt.Errorf("error resolving tools for %s: %w", id, err)
		}
		data["tools"] = resolvedTools
	}

	if doc.Data.Undo != nil {
		if doc.Data.Undo.ID == "" {
			doc.Data.Undo.ID = doc.Data.Undo.Name
		}
		data["undo"] = doc.Data.Undo
	}

	// Map Metadata
	if doc.Data.Metadata != nil {
		data["metadata"] = doc.Data.Metadata
	}
	if doc.Data.Timeout != "" {
		data["timeout"] = doc.Data.Timeout
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node data: %w", err)
	}

	return bytes, nil
}

// resolveTools recursively resolves polymorphic tool definitions (inline maps or import strings).
func (l *Loader) resolveTools(ctx context.Context, toolsRaw []any, visited map[string]bool) ([]domain.Tool, error) {
	if visited == nil {
		visited = make(map[string]bool)
	}

	// Use a map to handle overrides/deduplication by name.
	// "Local > Import" logic is handled by processing order:
	// We process value by value. Later values override earlier ones in the map.
	// So if imports come first, local (later) overrides.
	toolMap := make(map[string]domain.Tool)
	var toolNames []string // Keep track of order

	for _, item := range toolsRaw {
		switch v := item.(type) {
		case string:
			// Import Reference
			refID := trimExtension(v)
			if visited[refID] {
				return nil, fmt.Errorf("cycle detected in tool imports: %s", refID)
			}

			// DFS Cycle Detection: Mark
			visited[refID] = true

			doc, err := l.Repo.Get(ctx, refID)
			if err != nil {
				// Don't leak detail if not needed, but here it helps
				return nil, fmt.Errorf("failed to load imported tool library '%s': %w", refID, err)
			}

			importedTools, err := l.resolveTools(ctx, doc.Data.Tools, visited)

			// DFS Cycle Detection: Unmark (backtrack)
			delete(visited, refID)

			if err != nil {
				return nil, err
			}

			// Merge imported tools
			for _, t := range importedTools {
				if _, exists := toolMap[t.Name]; !exists {
					toolNames = append(toolNames, t.Name)
				}
				toolMap[t.Name] = t
			}

		case map[string]any, map[any]any:
			// Inline Definition
			var tool domain.Tool
			if err := mapstructure.Decode(v, &tool); err != nil {
				return nil, fmt.Errorf("failed to decode inline tool: %w", err)
			}
			if tool.Name == "" {
				return nil, fmt.Errorf("inline tool missing name")
			}
			if _, exists := toolMap[tool.Name]; !exists {
				toolNames = append(toolNames, tool.Name)
			}
			// Overwrite existing (Shadowing)
			toolMap[tool.Name] = tool

		default:
			return nil, fmt.Errorf("invalid tool definition type: %T", v)
		}
	}

	// Flatten result in order
	result := make([]domain.Tool, 0, len(toolNames))
	for _, name := range toolNames {
		result = append(result, toolMap[name])
	}

	return result, nil
}

// ListNodes lists all nodes in the repository.
func (l *Loader) ListNodes() ([]string, error) {
	ctx := context.Background()
	docs, err := l.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("loam list failed: %w", err)
	}

	ids := make([]string, len(docs))
	for i, doc := range docs {
		// Use the ID from metadata if available, otherwise filename ID
		rawID := doc.Data.ID
		if rawID == "" {
			rawID = doc.ID
		}
		ids[i] = trimExtension(rawID)
	}
	return ids, nil
}

func trimExtension(id string) string {
	ext := filepath.Ext(id)
	if ext != "" {
		return filepath.ToSlash(strings.TrimSuffix(id, ext))
	}
	return filepath.ToSlash(id)
}

// Watch implements ports.Watchable.
func (l *Loader) Watch(ctx context.Context) (<-chan string, error) {
	// Watch for all relevant files (recursive) using doublestar pattern supported by Loam/Doublestar
	// This avoids manual filtering loop.
	events, err := l.Repo.Watch(ctx, "**/*.{md,json,yaml,yml}")
	if err != nil {
		return nil, fmt.Errorf("failed to start loam watcher: %w", err)
	}

	ch := make(chan string, 1)

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-events:
				if !ok {
					return
				}
				// Pass the changed ID up the chain
				select {
				case ch <- evt.ID:
				default:
					// Drop event if channel is full (debounce)
				}
			}
		}
	}()

	return ch, nil
}
