# Running the MCP Server

Trellis supports the **Model Context Protocol (MCP)**, allowing AI agents (like Claude Desktop) to connect directly to your state machine.

This turns Trellis into a "Tool" for AI, enabling agents to:

1. **Navigate** your documentation flow.
2. **Render** the current state for context.
3. **Inspect** the graph structure to understand the document map.

## 1. Quick Start (Claude Desktop)

To use Trellis with Claude Desktop, you need to configure it as a local MCP server.

### Stdio Mode (Recommended)

This mode runs `trellis` as a subprocess of Claude Desktop.

1. Locate your Claude Desktop configuration file:
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

2. Add Trellis to the `mcpServers` section:

```json
{
  "mcpServers": {
    "trellis-tour": {
      "command": "go",
      "args": [
        "run",
        "github.com/aretw0/trellis/cmd/trellis",
        "mcp",
        "--dir",
        "C:/path/to/trellis/examples/tour" 
      ]
    }
  }
}
```

> **Note**: Replace `C:/path/to/trellis/...` with the absolute path to your flow directory.
> You can also compile the binary first (`go build -o trellis.exe ./cmd/trellis`) and point "command" to the executable.

1. Restart Claude Desktop.
2. You should see a "connection" icon. Ask Claude: *"Start the Trellis tour"*.

## 2. Remote / Network Mode (SSE)

If you are building a custom agent or want to debug the MCP traffic, you can run the server in **SSE (Server-Sent Events)** mode.

```bash
go run ./cmd/trellis mcp --dir ./examples/tour --transport sse --port 8080
```

*Or using the Makefile:*

```bash
make mcp-tour
```

The server will be available at:

- **SSE Endpoint**: `http://localhost:8080/sse`
- **Messages Endpoint**: `http://localhost:8080/messages`

You can connect to this using any MCP-compliant client (e.g., [MCP Inspector](https://github.com/modelcontextprotocol/inspector)).

## 3. Client Configuration Guide

### Visual Studio Code

Many VS Code extensions (like "MCP Inspector" or agentic extensions) support MCP. Configuration typically goes in `.vscode/mcp.json` (Workspace) or your User Settings.

```json
{
  "mcpServers": {
    "trellis-vscode": {
      "command": "go",
      "args": ["run", "./cmd/trellis", "mcp", "--dir", "./examples/tour"]
    }
  }
}
```

### Cursor

To use Trellis with [Cursor](https://cursor.com):

1. Go to **Settings > Cursor Settings > MCP**.
2. Click **Add new MCP server**.
3. Name: `trellis-cursor`
4. Type: `stdio`
5. Command: `go run ./cmd/trellis mcp --dir ./examples/tour`

*(Note: ensure you use absolute paths if the command fails to find files)*

### Antigravity (Agent)

Antigravity natively supports MCP. Add the server to your `mcp_config.json`:

```json
{
  "mcpServers": {
    "trellis": {
      "command": "go",
      "args": [
        "run", 
        "C:/path/to/trellis/cmd/trellis", 
        "mcp", 
        "--dir", 
        "C:/path/to/trellis/examples/tour"
      ]
      ]
    }
  }
}
```

### Gemini (Code Assist)

If your Gemini environment supports MCP (Model Context Protocol), it likely follows the standard JSON configuration for "Tools" or "Connectors". Use the same JSON structure as above (stdio mode) or connect via SSE URL if running remotely.

## 4. Available Tools & Resources

### Tools

- `render_state(node_id?)`: Returns the view (text actions) for a node.
- `navigate(node_id, input)`: Transitions to the next state.
- `get_graph()`: Returns the full JSON definition of the nodes.

### Resources

- `trellis://graph`: A read-only JSON resource containing the entire graph structure.

## 5. Debugging with MCP Inspector

You can use the official [MCP Inspector](https://github.com/modelcontextprotocol/inspector) to interactively test your server.

### Option A: Stdio Mode (Easiest)

The inspector spawns the Trellis process directly. This is the default.

```bash
make inspect-tour
```

*In the Inspector UI:* Select **Transport Type: Stdio**.

### Option B: SSE Mode (Advanced)

Useful for debugging remote connections or HTTP issues. Requires two terminals.

**Terminal 1 (Start Server):**

```bash
make mcp-tour
```

**Terminal 2 (Start Inspector):**

```bash
make inspect-tour-sse
```

*In the Inspector UI:* Select **Transport Type: SSE**. The URL should pre-fill as `http://localhost:8080/sse`.

## 6. Common Issues

If the connection fails:

- Check the local logs in Claude Desktop.
- Ensure the path to your flow directory is **absolute** and correct.
- Verify `trellis` compiles and runs via `trellis run` first.
