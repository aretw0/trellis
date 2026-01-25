.PHONY: all gen build serve test verify

# Default target
all: gen build

# Generate Go code from OpenAPI spec using oapi-codegen
gen:
	go generate ./internal/adapters/http

# Build the Trellis CLI
build:
	go build -o trellis.exe ./cmd/trellis

# Run the Stateless Server in dev mode (requires `tour` example)
serve-tour: gen
	go run ./cmd/trellis serve --dir ./examples/tour --port 8080

# Run the MCP Server in SSE mode (requires `tour` example)
mcp-tour:
	go run ./cmd/trellis mcp --dir ./examples/tour --transport sse --port 8080

# Run the MCP Inspector against the Tour example (using Stdio)
inspect-tour:
	npx @modelcontextprotocol/inspector go run ./cmd/trellis mcp --dir ./examples/tour

# Run the MCP Inspector against a running SSE server (requires 'make mcp-tour' in another terminal)
inspect-tour-sse:
	npx @modelcontextprotocol/inspector --server-url http://localhost:8080/sse


# Run all tests
test:
	go test ./...

# Run vet tool in all files
vet:
	go vet ./...

# Verify server endpoints (requires server running in another terminal)
verify:
	curl.exe -X POST http://localhost:8080/render -H "Content-Type: application/json" -d "{\"current_node_id\": \"start\"}"


# Run local Go documentation server (pkgsite)
serve-docs:
	go tool godoc -http=:6060
