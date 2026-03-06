package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultInterpolator(t *testing.T) {
	ctx := context.Background()

	t.Run("FastPath_NoTemplateTokens", func(t *testing.T) {
		input := "Hello, world!"
		result, err := runtime.DefaultInterpolator(ctx, input, nil)
		require.NoError(t, err)
		assert.Equal(t, input, result, "should return the string unchanged without parsing")
	})

	t.Run("SimpleKeyInterpolation", func(t *testing.T) {
		data := map[string]any{"username": "Alice"}
		result, err := runtime.DefaultInterpolator(ctx, "Hello, {{ .username }}!", data)
		require.NoError(t, err)
		assert.Equal(t, "Hello, Alice!", result)
	})

	t.Run("MissingKey_ReturnsNoValue", func(t *testing.T) {
		data := map[string]any{}
		result, err := runtime.DefaultInterpolator(ctx, "Hello, {{ .username }}!", data)
		require.NoError(t, err)
		// In Go's text/template, absent map keys always render as "<no value>".
		// Use {{ default "fallback" .key }} to provide explicit fallbacks.
		assert.Equal(t, "Hello, <no value>!", result)
	})

	t.Run("FuncMap_Default_MissingKey", func(t *testing.T) {
		data := map[string]any{}
		result, err := runtime.DefaultInterpolator(ctx, `{{ default "visitante" .username }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "visitante", result)
	})

	t.Run("FuncMap_Default_PresentKey", func(t *testing.T) {
		data := map[string]any{"username": "Alice"}
		result, err := runtime.DefaultInterpolator(ctx, `{{ default "visitante" .username }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "Alice", result)
	})

	t.Run("FuncMap_Default_EmptyString_UsesDefault", func(t *testing.T) {
		// An empty string is zero — default should kick in.
		data := map[string]any{"username": ""}
		result, err := runtime.DefaultInterpolator(ctx, `{{ default "guest" .username }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "guest", result)
	})

	t.Run("FuncMap_Coalesce_ReturnsFirstNonZero", func(t *testing.T) {
		data := map[string]any{"a": "", "b": "second", "c": "third"}
		result, err := runtime.DefaultInterpolator(ctx, `{{ coalesce .a .b .c }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "second", result)
	})

	t.Run("FuncMap_Coalesce_AllZero_ReturnsNoValue", func(t *testing.T) {
		data := map[string]any{"a": "", "b": ""}
		result, err := runtime.DefaultInterpolator(ctx, `{{ coalesce .a .b }}`, data)
		require.NoError(t, err)
		// coalesce returns nil when all values are zero.
		// Go templates render nil as "<no value>" for map key access.
		assert.Equal(t, "<no value>", result)
	})

	t.Run("FuncMap_ToJson_Map", func(t *testing.T) {
		data := map[string]any{
			"tool_result": map[string]any{"id": "call-1", "result": "ok"},
		}
		result, err := runtime.DefaultInterpolator(ctx, `{{ toJson .tool_result }}`, data)
		require.NoError(t, err)
		assert.JSONEq(t, `{"id":"call-1","result":"ok"}`, result)
	})

	t.Run("NativeFunc_Index_MapAccess", func(t *testing.T) {
		data := map[string]any{
			"config": map[string]any{"env": "production"},
		}
		// index is a builtin Go template function — no manual registration needed.
		result, err := runtime.DefaultInterpolator(ctx, `{{ index .config "env" }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "production", result)
	})

	t.Run("InvalidTemplate_ReturnsError", func(t *testing.T) {
		_, err := runtime.DefaultInterpolator(ctx, "{{ .unclosed", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid template")
	})

	t.Run("TextTemplate_DoesNotEscapeHTML", func(t *testing.T) {
		data := map[string]any{"content": "<b>bold</b>"}
		result, err := runtime.DefaultInterpolator(ctx, `{{ .content }}`, data)
		require.NoError(t, err)
		// text/template must NOT escape HTML — raw string expected.
		assert.Equal(t, "<b>bold</b>", result)
	})

	t.Run("Conditional_TruthyValue", func(t *testing.T) {
		data := map[string]any{"user_input": "hello"}
		result, err := runtime.DefaultInterpolator(ctx, `{{ if .user_input }}set{{ else }}not set{{ end }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "set", result)
	})

	t.Run("Conditional_MissingKey_IsFalsy", func(t *testing.T) {
		data := map[string]any{}
		result, err := runtime.DefaultInterpolator(ctx, `{{ if .user_input }}set{{ else }}not set{{ end }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "not set", result)
	})

	t.Run("ToolResult_FlatMap_AccessibleFields", func(t *testing.T) {
		// Simulates the context shape produced by navigation.go after a successful tool call payload (Map flattened).
		data := map[string]any{
			"tool_result": map[string]any{
				"_id":    "call-abc",
				"status": 200,
				"body":   "success-payload",
			},
		}
		result, err := runtime.DefaultInterpolator(ctx, `id={{ .tool_result._id }} status={{ .tool_result.status }} body={{ .tool_result.body }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "id=call-abc status=200 body=success-payload", result)
	})

	t.Run("ToolResult_Scalar_AccessibleAsResult", func(t *testing.T) {
		// Simulates the context shape produced by navigation.go for a scalar tool result.
		data := map[string]any{
			"tool_result": map[string]any{
				"_id":    "call-xyz",
				"result": "OK",
			},
		}
		result, err := runtime.DefaultInterpolator(ctx, `id={{ .tool_result._id }} result={{ .tool_result.result }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "id=call-xyz result=OK", result)
	})

}

func TestHTMLInterpolator(t *testing.T) {
	ctx := context.Background()

	t.Run("EscapesHTMLCharacters", func(t *testing.T) {
		data := map[string]any{"content": "<b>bold</b>"}
		result, err := runtime.HTMLInterpolator(ctx, `{{ .content }}`, data)
		require.NoError(t, err)
		// html/template MUST escape HTML.
		assert.Equal(t, "&lt;b&gt;bold&lt;/b&gt;", result)
	})

	t.Run("FuncMap_Default_WorksSameAsDefault", func(t *testing.T) {
		data := map[string]any{}
		result, err := runtime.HTMLInterpolator(ctx, `{{ default "N/A" .missing }}`, data)
		require.NoError(t, err)
		assert.Equal(t, "N/A", result)
	})

	t.Run("FastPath_NoTemplateTokens", func(t *testing.T) {
		input := "plain text"
		result, err := runtime.HTMLInterpolator(ctx, input, nil)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})
}
