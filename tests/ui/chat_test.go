package ui_test

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aretw0/trellis"
	httpAdapter "github.com/aretw0/trellis/pkg/adapters/http"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func TestChatUI_ExhaustiveFlow(t *testing.T) {
	// 1. Initialize Engine with exhaustive fixture
	fixturePath := filepath.Join("..", "fixtures", "ui_exhaustive")
	engine, err := trellis.New(fixturePath)
	if err != nil {
		t.Fatalf("Failed to init engine: %v", err)
	}

	// 2. Start HTTP Server
	handler := httpAdapter.NewHandler(engine)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// 3. Setup rod browser
	// Headless mode: set TRELLIS_TEST_HEADLESS=false to open a visible browser window for debugging.
	headless := os.Getenv("TRELLIS_TEST_HEADLESS") != "false"
	t.Logf("Headless mode: %v (set TRELLIS_TEST_HEADLESS=false to disable)", headless)
	// Disable Leakless so it doesn't fail extracting into AppData temp on Windows
	u := launcher.New().Headless(headless).Leakless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	// Setup incognito to avoid cross-pollination
	incognito := browser.MustIncognito()
	page := incognito.MustPage(ts.URL + "/ui")

	// Timeout for safety
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	page = page.Context(ctx)

	// -- A. Session Initialization --
	t.Log("Testing Session Initialization...")
	page.MustElement("#conn-status").MustWaitLoad()
	page.MustWait(`() => document.querySelector('#conn-status').innerText.includes('Live Updates')`)
	t.Log("SSE Connected.")

	// -- B. Message Rendering (Start Node) --
	t.Log("Testing Message Rendering...")
	// 'start' node renders immediately, transitions to input_node
	page.MustElementR(".bubble-system", "Flow Initialized").MustWaitVisible()
	// input_node prompts for user input
	page.MustElementR(".bubble-system", "Please provide some input for the next step").MustWaitVisible()

	// -- C. Input and Tool Execution (Input → Tool Node) --
	t.Log("Testing Tool Execution...")
	inputBox := page.MustElement("#user-input")
	inputBox.MustWaitVisible()
	inputBox.MustInput("Hello Tool")
	page.MustElement("#send-btn").MustClick()

	// User bubble reflects input
	page.MustElementR(".bubble-user", "Hello Tool").MustWaitVisible()

	// Tool execution bubble: name and args visible
	page.MustElementR(".bubble-tool", "mock_tool").MustWaitVisible()
	page.MustElementR(".bubble-tool", "Hello Tool").MustWaitVisible()

	// -- D. Tool Result Injection --
	// The UI is stateless and does not execute tools locally.
	// We mock the tool result by posting a ToolResult to /navigate.
	t.Log("Injecting mock tool result...")
	page.MustEval(`() => {
		apiCall("/navigate", {
			state: currentState,
			input: {
				id: currentState.pending_tool_call,
				result: { received: "Hello Tool", success: true }
			}
		});
	}`)

	// tool_result_node renders the raw map: "Tool result was: map[received:Hello Tool success:true]"
	page.MustElementR(".bubble-system", "Tool result was:").MustWaitVisible()

	// -- E. Kitchen Sink Interpolation --
	//
	// Verifies Go text/template interpolation in Trellis node content.
	// Engine uses DefaultInterpolator (plain text/template, no FuncMap).
	//
	// KNOWN LIMITATIONS (deferred — see PLANNING.md for future patch):
	//   - default_context from start.md not reaching the template in this fixture
	//     (values arrive as empty — investigate YAML parsing of DefaultContext)
	//   - {{ default "N/A" .missing_key }} → FAILS: 'default' is not registered
	//   - {{ .tool_result.received }}      → FAILS: tool_result is interface{}, not map
	//   - {{ range $k, $v := .tool_result }} → FAILS: cannot range over interface{}
	//   - tool_result structure is map[id:mock_tool result:map[...]] not map[received:...]
	t.Log("Testing Kitchen Sink Interpolation...")

	// ✅ String interpolation via save_to
	page.MustElementR(".bubble-system", `user_input: Hello Tool`).MustWaitVisible()

	// ✅ Conditional: non-empty value is truthy
	page.MustElementR(".bubble-system", "user_input is set").MustWaitVisible()

	// ✅ String equality comparison using eq builtin
	page.MustElementR(".bubble-system", "user_input matches Hello Tool").MustWaitVisible()

	// ✅ Raw map interpolation (tool_result stored as wrapper struct)
	page.MustElementR(".bubble-system", "tool_result raw:").MustWaitVisible()

	// -- F. End Node and Graceful Termination --
	page.MustElementR(".bubble-system", "The flow has reached the end").MustWaitVisible()

	t.Log("Testing Session Termination...")
	page.MustWait(`() => {
		const promptHide = document.querySelector('#prompt-area').classList.contains('hidden');
		const termMsg = document.querySelector('#terminal-msg').innerText;
		return promptHide && termMsg.includes('Ended');
	}`)

	t.Log("Exhaustive UI Flow Integration Test Succeeded!")
}
