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

	transitions := l.buildTransitions(doc.Data)
	data, err := l.buildBaseNodeData(doc.ID, doc.Data, doc.Content, transitions)
	if err != nil {
		return nil, err
	}
	l.applyInputConfig(doc.Data, data)

	if err := l.applyToolConfig(ctx, id, doc.Data, data); err != nil {
		return nil, err
	}

	l.applyMetadataAndTimeout(doc.Data, data)

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node data: %w", err)
	}

	return bytes, nil
}

func (l *Loader) buildTransitions(meta NodeMetadata) []domain.Transition {
	totalTransitions := len(meta.Transitions) + len(meta.Options)
	transitions := make([]domain.Transition, 0, totalTransitions)

	convert := func(lt LoaderTransition) domain.Transition {
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
		if condition == "" && lt.Text != "" {
			condition = fmt.Sprintf("input == '%s'", strings.ReplaceAll(lt.Text, "'", "\\'"))
		}

		return domain.Transition{
			FromNodeID: from,
			ToNodeID:   to,
			Condition:  condition,
		}
	}

	for _, opt := range meta.Options {
		transitions = append(transitions, convert(opt))
	}
	for _, lt := range meta.Transitions {
		transitions = append(transitions, convert(lt))
	}
	if meta.To != "" {
		transitions = append(transitions, domain.Transition{ToNodeID: meta.To})
	}

	return transitions
}

func (l *Loader) buildBaseNodeData(docID string, meta NodeMetadata, content string, transitions []domain.Transition) (map[string]any, error) {
	data := make(map[string]any)

	rawID := meta.ID
	if rawID == "" {
		rawID = docID
	}
	data["id"] = trimExtension(rawID)

	data["type"] = meta.Type
	if meta.Type == "" {
		data["type"] = domain.NodeTypeText
	}

	if meta.Wait {
		data["wait"] = meta.Wait
	}
	if meta.OnError != "" {
		data["on_error"] = meta.OnError
	}
	if meta.OnDenied != "" {
		data["on_denied"] = meta.OnDenied
	}
	if len(meta.OnSignal) > 0 {
		data["on_signal"] = meta.OnSignal
	}

	l.applySignalSugar(meta, data)

	if meta.SaveTo != "" {
		data["save_to"] = meta.SaveTo
	}
	if len(meta.RequiredContext) > 0 {
		data["required_context"] = meta.RequiredContext
	}
	if len(meta.DefaultContext) > 0 {
		data["default_context"] = meta.DefaultContext
	}
	if len(meta.ContextSchema) > 0 {
		normalized, err := normalizeContextSchema(meta.ContextSchema)
		if err != nil {
			return nil, err
		}
		data["context_schema"] = normalized
	}

	data["transitions"] = transitions
	data["content"] = []byte(content)

	return data, nil
}

func normalizeContextSchema(raw map[string]any) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	normalized := make(map[string]string, len(raw))
	for key, value := range raw {
		typeStr, err := formatSchemaType(value)
		if err != nil {
			return nil, fmt.Errorf("context_schema.%s: %w", key, err)
		}
		normalized[key] = typeStr
	}

	return normalized, nil
}

func formatSchemaType(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []any:
		if len(v) != 1 {
			return "", fmt.Errorf("expected single element list for slice type")
		}
		inner, err := formatSchemaType(v[0])
		if err != nil {
			return "", err
		}
		return "[" + inner + "]", nil
	case []string:
		if len(v) != 1 {
			return "", fmt.Errorf("expected single element list for slice type")
		}
		return "[" + v[0] + "]", nil
	default:
		return "", fmt.Errorf("expected string or list, got %T", value)
	}
}

func (l *Loader) applySignalSugar(meta NodeMetadata, data map[string]any) {
	if meta.OnTimeout != "" {
		signals := l.ensureSignalMap(data)
		signals[domain.SignalTimeout] = meta.OnTimeout
	}
	if meta.OnInterrupt != "" {
		signals := l.ensureSignalMap(data)
		signals[domain.SignalInterrupt] = meta.OnInterrupt
	}
}

func (l *Loader) ensureSignalMap(data map[string]any) map[string]string {
	signals, ok := data["on_signal"].(map[string]string)
	if !ok {
		signals = make(map[string]string)
		data["on_signal"] = signals
	}
	return signals
}

func (l *Loader) applyInputConfig(meta NodeMetadata, data map[string]any) {
	inferredOptions := make([]string, 0)
	for _, opt := range meta.Options {
		if opt.Text != "" {
			inferredOptions = append(inferredOptions, opt.Text)
		}
	}

	inputType := meta.InputType
	inputOptions := meta.InputOptions

	if len(inferredOptions) > 0 {
		if len(inputOptions) == 0 {
			inputOptions = inferredOptions
		}
		if inputType == "" {
			inputType = string(domain.InputChoice)
		}
	}

	if inputType == string(domain.InputConfirm) && len(inputOptions) == 0 {
		inputOptions = []string{"yes", "no"}
	}

	if inputType != "" {
		data["input_type"] = inputType
		data["input_options"] = inputOptions
		data["input_default"] = meta.InputDefault
	}
}

func (l *Loader) applyToolConfig(ctx context.Context, nodeID string, meta NodeMetadata, data map[string]any) error {
	var toolCall *LoaderToolCall
	if meta.Do != nil {
		toolCall = meta.Do
	} else if meta.ToolCall != nil {
		toolCall = meta.ToolCall
	}

	if toolCall != nil {
		domainCall := convertToolCall(toolCall)
		if domainCall.ID == "" {
			domainCall.ID = domainCall.Name
		}
		data["do"] = domainCall
	}

	if len(meta.Tools) > 0 {
		resolvedTools, err := l.resolveTools(ctx, meta.Tools, nil)
		if err != nil {
			return fmt.Errorf("error resolving tools for %s: %w", nodeID, err)
		}
		data["tools"] = resolvedTools
	}

	if meta.Undo != nil {
		domainUndo := convertToolCall(meta.Undo)
		if domainUndo.ID == "" {
			domainUndo.ID = domainUndo.Name
		}
		data["undo"] = domainUndo
	}

	return nil
}

func (l *Loader) applyMetadataAndTimeout(meta NodeMetadata, data map[string]any) {
	if meta.Metadata != nil {
		data["metadata"] = flattenMetadata(meta.Metadata)
	}
	if meta.Timeout != "" {
		data["timeout"] = meta.Timeout
	}
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

	seen := make(map[string]string)
	ids := make([]string, 0, len(docs))

	for _, doc := range docs {
		// Use the ID from metadata if available, otherwise filename ID
		rawID := doc.Data.ID
		if rawID == "" {
			rawID = doc.ID
		}
		id := trimExtension(rawID)

		// Collision Detection
		if existingPath, ok := seen[id]; ok {
			// doc.ID is usually the filepath in Loam (or relative path)
			return nil, fmt.Errorf("collision detected: ID '%s' is defined in both '%s' and '%s'", id, existingPath, doc.ID)
		}
		seen[id] = doc.ID
		ids = append(ids, id)
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
				// Loam v0.10.9 provides its own resilient debounce via lifecycle.
				// Pass the changed ID up the chain, respecting context cancellation.
				select {
				case ch <- evt.ID:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// flattenMetadata converts a nested map[string]any into a flat map[string]string
// using dot notation (or dash for compatibility) for keys.
func flattenMetadata(src map[string]any) map[string]string {
	res := make(map[string]string)
	var visit func(prefix string, v any)

	visit = func(prefix string, v any) {
		switch val := v.(type) {
		case map[string]any:
			for k, sub := range val {
				fullKey := k
				if prefix != "" {
					// Use '-' separator for x-exec compatibility (x-exec-command)
					// Generally dot is better, but our hypothesis was x-exec-command.
					// Let's check prefix. If prefix is "x-exec", we want "x-exec-command".
					fullKey = prefix + "-" + k
				}
				visit(fullKey, sub)
			}
		case map[interface{}]interface{}: // YAML often decodes to this
			for k, sub := range val {
				strKey := fmt.Sprintf("%v", k)
				fullKey := strKey
				if prefix != "" {
					fullKey = prefix + "-" + strKey
				}
				visit(fullKey, sub)
			}
		case []any:
			// Join arrays as space-separated strings (useful for args)
			var parts []string
			for _, item := range val {
				parts = append(parts, fmt.Sprintf("%v", item))
			}
			res[prefix] = strings.Join(parts, " ")
		default:
			if prefix != "" {
				res[prefix] = fmt.Sprintf("%v", val)
			}
		}
	}

	for k, v := range src {
		visit(k, v)
	}
	return res
}

func convertToolCall(src *LoaderToolCall) *domain.ToolCall {
	if src == nil {
		return nil
	}
	return &domain.ToolCall{
		ID:             src.ID,
		Name:           src.Name,
		Args:           src.Args,
		Metadata:       flattenMetadata(src.Metadata),
		IdempotencyKey: src.IdempotencyKey,
	}
}
