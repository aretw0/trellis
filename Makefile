.PHONY: all gen build test vet tidy serve-docs serve-tour mcp-tour inspect-tour inspect-tour-sse verify work-on-lifecycle work-on-loam work-on-procio work-off-lifecycle work-off-loam work-off-procio work-off-all

# Default target
all: gen build

# Generate Go code from OpenAPI spec using oapi-codegen
gen:
	go generate ./internal/adapters/http

# Build the Trellis CLI
build:
	go build -o trellis.exe ./cmd/trellis

# Run all tests
# Note: -race is mandatory for verifying behavioral logic and concurrency safety.
test:
	go test -race -timeout 90s ./...

# Run vet tool in all files
vet:
	go vet ./...

# Ensure dependencies are clean
tidy:
	go mod tidy

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

# Enable local development mode for lifecycle by creating/updating go.work
# Usage: make work-on-lifecycle [WORK_PATH=../lifecycle]
work-on-lifecycle:
	@echo "Enabling local lifecycle..."
	@if not exist go.work ( echo "Initializing go.work..." & go work init . )
	@if "$(WORK_PATH)"=="" ( go work use ../lifecycle ) else ( go work use $(WORK_PATH) )

# Enable local development mode for loam by creating/updating go.work
# Usage: make work-on-loam [WORK_PATH=../loam]
work-on-loam:
	@echo "Enabling local loam..."
	@if not exist go.work ( echo "Initializing go.work..." & go work init . )
	@if "$(WORK_PATH)"=="" ( go work use ../loam ) else ( go work use $(WORK_PATH) )

# Enable local development mode for procio by creating/updating go.work
# Usage: make work-on-procio [WORK_PATH=../procio]
work-on-procio:
	@echo "Enabling local procio..."
	@if not exist go.work ( echo "Initializing go.work..." & go work init . )
	@if "$(WORK_PATH)"=="" ( go work use ../procio ) else ( go work use $(WORK_PATH) )

# Disable local lifecycle (remove from go.work)
# Usage: make work-off-lifecycle [WORK_PATH=../lifecycle]
work-off-lifecycle:
	@echo "Disabling local lifecycle..."
	@if exist go.work ( \
		if "$(WORK_PATH)"=="" ( go work edit -dropuse ../lifecycle ) else ( go work edit -dropuse $(WORK_PATH) ) \
	)

# Disable local loam (remove from go.work)
# Usage: make work-off-loam [WORK_PATH=../loam]
work-off-loam:
	@echo "Disabling local loam..."
	@if exist go.work ( \
		if "$(WORK_PATH)"=="" ( go work edit -dropuse ../loam ) else ( go work edit -dropuse $(WORK_PATH) ) \
	)

# Disable local procio (remove from go.work)
# Usage: make work-off-procio [WORK_PATH=../procio]
work-off-procio:
	@echo "Disabling local procio..."
	@if exist go.work ( \
		if "$(WORK_PATH)"=="" ( go work edit -dropuse ../procio ) else ( go work edit -dropuse $(WORK_PATH) ) \
	)

# Disable local development mode by removing go.work (nuclear option)
work-off-all:
	@echo "Disabling local workspace mode..."
	@if exist go.work ( del go.work )