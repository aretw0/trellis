package process

import (
	"context"
	"runtime"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Execute(t *testing.T) {
	// Setup: Define a command that works on the OS
	cmdName := "echo"
	args := []string{"hello"}
	if runtime.GOOS == "windows" {
		// On Windows, echo is a shell builtin, so we use "cmd /c echo" or a simple executable.
		// Go's exec.LookPath might not find "echo".
		// We can use "go version" as a safe cross-platform command.
		cmdName = "go"
		args = []string{"version"}
	}

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
			testArgs = []string{"/c", "echo %TRELLIS_ARG_MSG%"}
		} else {
			testCmd = "sh"
			testArgs = []string{"-c", "echo $TRELLIS_ARG_MSG"}
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
		assert.False(t, result.IsError)
		assert.Contains(t, result.Result.(string), "SecretMessage")
	})
}
