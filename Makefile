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

# Run all tests
test:
	go test ./...

# Verify server endpoints (requires server running in another terminal)
verify:
	curl.exe -X POST http://localhost:8080/render -H "Content-Type: application/json" -d "{\"current_node_id\": \"start\"}"
