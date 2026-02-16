.PHONY: all gen build test coverage lint vet tidy serve-docs serve-tour mcp-tour inspect-tour inspect-tour-sse verify work-on-lifecycle work-on-loam work-on-procio work-on-introspection work-off-lifecycle work-off-loam work-off-procio work-off-introspection work-off-all

# Go parameters
GOCMD=go
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run
GOWORK=$(GOCMD) work
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOINSTALL=$(GOCMD) install
BINARY_NAME=loam
BINARY_UNIX=$(BINARY_NAME)

# --- OS Detection & Command Abstraction ---
ifeq ($(OS),Windows_NT)
BINARY := $(BINARY_NAME).exe
RM := del /F /Q
# Windows needs backslashes for 'go work edit -dropuse' to match go.work content
DROP_WORK = if exist go.work ( $(GOWORK) edit -dropuse $(subst /,\,$(1)) )
INIT_WORK = if not exist go.work ( echo "Initializing go.work..." & $(GOWORK) init . )
else
BINARY := $(BINARY_UNIX)
RM := rm -f
# Linux/macOS uses forward slashes
DROP_WORK = [ -f go.work ] && $(GOWORK) edit -dropuse $(1)
INIT_WORK = [ -f go.work ] || ( echo "Initializing go.work..." && $(GOWORK) init . )
endif

# --- OS Detection & Command Abstraction ---
ifeq ($(OS),Windows_NT)
BINARY := trellis.exe
RM := del /F /Q
CURL := curl.exe
# Windows needs backslashes for 'go work edit -dropuse' to match go.work content
DROP_WORK = if exist go.work ( $(GOWORK) edit -dropuse $(subst /,\,$(1)) )
INIT_WORK = if not exist go.work ( echo "Initializing go.work..." & $(GOWORK) init . )
else
BINARY := trellis
RM := rm -f
CURL := curl
# Linux/macOS uses forward slashes
DROP_WORK = [ -f go.work ] && $(GOWORK) edit -dropuse $(1)
INIT_WORK = [ -f go.work ] || ( echo "Initializing go.work..." && $(GOWORK) init . )
endif

# Default target
all: gen build

# Generate Go code from OpenAPI spec using oapi-codegen
gen:
	go generate ./internal/adapters/http

# Build the binary for the current platform
build:
	@echo "Building for $(GOOS)/$(GOARCH)..."
	$(GOBUILD) -v -o $(BINARY_NAME) ./cmd/trellis

# Run all tests
# Note: -race is mandatory for verifying behavioral logic and concurrency safety.
test:
	@echo "Running tests..."
	$(GOTEST) -race -timeout 90s ./...

# Run coverage tests
coverage:
	$(GOTEST) -race -timeout 90s -coverprofile="coverage.out" ./...
	$(GOTOOL) cover -func="coverage.out"

# Run ineffassign to detect ineffectual assignments
lint:
	$(GORUN) github.com/gordonklaus/ineffassign@latest ./...

# Run vet tool in all files
vet:
	go vet ./...

# Ensure dependencies are clean
tidy:
	$(GOMOD) tidy

# Run local Go documentation server (pkgsite)
serve-docs:
	$(GOTOOL) godoc -http=:6060

# Run the Stateless Server in dev mode (requires `tour` example)
serve-tour: gen
	$(GORUN) ./cmd/trellis serve --dir ./examples/tour --port 8080

# Run the MCP Server in SSE mode (requires `tour` example)
mcp-tour:
	$(GORUN) ./cmd/trellis mcp --dir ./examples/tour --transport sse --port 8080

# Run the MCP Inspector against the Tour example (using Stdio)
inspect-tour:
	npx @modelcontextprotocol/inspector $(GORUN) ./cmd/trellis mcp --dir ./examples/tour

# Run the MCP Inspector against a running SSE server (requires 'make mcp-tour' in another terminal)
inspect-tour-sse:
	npx @modelcontextprotocol/inspector --server-url http://localhost:8080/sse

# Verify server endpoints (requires server running in another terminal)
verify:
	$(CURL) -X POST http://localhost:8080/render -H "Content-Type: application/json" -d "{\"current_node_id\": \"start\"}"

# --- Dependency Management (Dev vs Prod) ---

# Helper to get the correct path (uses WORK_PATH if provided, else default)
GET_PATH = $(if $(WORK_PATH),$(WORK_PATH),$(1))

# Enable local development mode for lifecycle
# Usage: make work-on-lifecycle [WORK_PATH=../lifecycle]
work-on-lifecycle:
	@echo "Enabling local lifecycle..."
	@$(INIT_WORK)
	@$(GOWORK) use $(call GET_PATH,../lifecycle)

# Enable local development mode for loam
# Usage: make work-on-loam [WORK_PATH=../loam]
work-on-loam:
	@echo "Enabling local loam..."
	@$(INIT_WORK)
	@$(GOWORK) use $(call GET_PATH,../loam)

# Enable local development mode for procio
# Usage: make work-on-procio [WORK_PATH=../procio]
work-on-procio:
	@echo "Enabling local procio..."
	@$(INIT_WORK)
	@$(GOWORK) use $(call GET_PATH,../procio)

# Enable local development mode for introspection
# Usage: make work-on-introspection [WORK_PATH=../introspection]
work-on-introspection:
	@echo "Enabling local introspection..."
	@$(INIT_WORK)
	@$(GOWORK) use $(call GET_PATH,../introspection)

# Disable local lifecycle
# Usage: make work-off-lifecycle [WORK_PATH=../lifecycle]
work-off-lifecycle:
	@echo "Disabling local lifecycle..."
	@$(call DROP_WORK,$(call GET_PATH,../lifecycle))

# Disable local loam
# Usage: make work-off-loam [WORK_PATH=../loam]
work-off-loam:
	@echo "Disabling local loam..."
	@$(call DROP_WORK,$(call GET_PATH,../loam))

# Disable local procio
# Usage: make work-off-procio [WORK_PATH=../procio]
work-off-procio:
	@echo "Disabling local procio..."
	@$(call DROP_WORK,$(call GET_PATH,../procio))

# Disable local introspection
# Usage: make work-off-introspection [WORK_PATH=../introspection]
work-off-introspection:
	@echo "Disabling local introspection..."
	@$(call DROP_WORK,$(call GET_PATH,../introspection))

# Disable local development mode by removing go.work (nuclear option)
work-off-all:
	@echo "Disabling local workspace mode..."
	@$(RM) go.work
