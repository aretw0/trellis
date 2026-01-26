.PHONY: all gen build test vet serve-docs serve-tour mcp-tour inspect-tour inspect-tour-sse verify use-local-lifecycle use-local-loam use-local-all use-pub-lifecycle use-pub-loam use-pub-all

# Default target
all: gen build

# Generate Go code from OpenAPI spec using oapi-codegen
gen:
	go generate ./internal/adapters/http

# Build the Trellis CLI
build:
	go build -o trellis.exe ./cmd/trellis

# Run all tests
test:
	go test ./...

# Run vet tool in all files
vet:
	go vet ./...

# Run local Go documentation server (pkgsite)
serve-docs:
	go tool godoc -http=:6060

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

# Verify server endpoints (requires server running in another terminal)
verify:
	curl.exe -X POST http://localhost:8080/render -H "Content-Type: application/json" -d "{\"current_node_id\": \"start\"}"

# --- Dependency Management (Dev vs Prod) ---

# Switch specific dependencies to local version
use-local-lifecycle:
	@echo "Switching lifecycle to local..."
	@go mod edit -replace github.com/aretw0/lifecycle=../lifecycle
	@go mod tidy

use-local-loam:
	@echo "Switching loam to local..."
	@go mod edit -replace github.com/aretw0/loam=../loam
	@go mod tidy

# Switch ALL known aretw0 dependencies to default local paths (siblings)
use-local-all:
	@echo "Switching all aretw0 deps to local (siblings)..."
	@go mod edit -replace github.com/aretw0/lifecycle=../lifecycle
	@go mod edit -replace github.com/aretw0/loam=../loam
	@go mod tidy

# Revert specific dependencies to published version
use-pub-lifecycle:
	@echo "Reverting lifecycle to published..."
	@go mod edit -dropreplace github.com/aretw0/lifecycle
	@go mod tidy

use-pub-loam:
	@echo "Reverting loam to published..."
	@go mod edit -dropreplace github.com/aretw0/loam
	@go mod tidy

# Revert ALL aretw0 dependencies to published versions
use-pub-all:
	@echo "Reverting all aretw0 deps to published..."
	@go mod edit -dropreplace github.com/aretw0/lifecycle
	@go mod edit -dropreplace github.com/aretw0/loam
	@go mod tidy
