package process

import (
	"context"
	"encoding/json"
	"runtime"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Execute(t *testing.T) {
	// Setup: Define a command that works on the OS.
	// We use "go version" as a safe cross-platform command.
	cmdName := "go"
	args := []string{"version"}

	runner := NewRunner()
	runner.Register("check_version", cmdName, args...)

	t.Run("Executes Registered Command", func(t *testing.T) {
		toolCall := domain.ToolCall{
			ID:   "call_1",
			Name: "check_version",
		}

		result, err := runner.Execute(context.Background(), toolCall)
		assert.NoError(t, err)
		assert.False(t, result.IsError)
		// Output should contain "go version" or equivalent
		assert.Contains(t, result.Result.(string), "go version")
	})

	t.Run("Fails For Unregistered Command", func(t *testing.T) {
		toolCall := domain.ToolCall{
			ID:   "call_2",
			Name: "hacker_script",
		}

		result, err := runner.Execute(context.Background(), toolCall)
		assert.NoError(t, err) // Should not return go error, but ToolResult error
		assert.True(t, result.IsError)
		assert.Contains(t, result.Error, "not registered")
	})

	t.Run("Passes Arguments via Env Vars", func(t *testing.T) {
		// For this test, we need a script that prints env vars.
		// We can use "go run" on a small snippet or use OS logic.
		// Linux: sh -c 'echo $TRELLIS_ARG_TEST'
		// Windows: cmd /c echo %TRELLIS_ARG_TEST%

		var testCmd string
		var testArgs []string

		if runtime.GOOS == "windows" {
			testCmd = "cmd"
			testArgs = []string{"/c", "echo %TRELLIS_ARGS%"}
		} else {
			testCmd = "sh"
			testArgs = []string{"-c", "echo $TRELLIS_ARGS"}
		}

		runner.Register("echo_env", testCmd, testArgs...)

		toolCall := domain.ToolCall{
			ID:   "call_3",
			Name: "echo_env",
			Args: map[string]any{
				"msg": "SecretMessage",
			},
		}

		result, err := runner.Execute(context.Background(), toolCall)
		assert.NoError(t, err)
		assert.False(t, result.IsError, "Tool result should not be an error: %v", result.Error)
		assert.NotNil(t, result.Result, "Result should not be nil")

		var output string
		switch v := result.Result.(type) {
		case string:
			output = v
		default:
			outputBytes, _ := json.Marshal(v)
			output = string(outputBytes)
		}
		assert.Contains(t, output, "\"msg\":\"SecretMessage\"")
	})
}
