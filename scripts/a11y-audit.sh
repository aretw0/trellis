#!/bin/bash

# Trellis Inspector A11y & Quality Audit
# Requires: pa11y, lighthouse (npm install -g pa11y lighthouse)

set -e

PORT=8081
URL="http://localhost:$PORT"
EXAMPLE_DIR="./examples/tour"

echo "Starting Trellis server for audit..."
go run ./cmd/trellis serve --dir "$EXAMPLE_DIR" --port "$PORT" &
SERVER_PID=$!

# Cleanup on exit
trap "kill $SERVER_PID" EXIT

# Wait for server to be ready
echo "Waiting for server to be ready..."
for i in {1..10}; do
  if curl -s "$URL" > /dev/null; then
    break
  fi
  sleep 1
done

echo "--- Running pa11y WCAG 2.1 AA Audit ---"
pa11y --standard WCAG2AA "$URL"

echo "--- Running Lighthouse Accessibility Audit ---"
# We use --chrome-flags="--headless" for CI compliance
lighthouse "$URL" --only-categories=accessibility --output=json --chrome-flags="--headless" > lighthouse-results.json

# Check Lighthouse score (threshold 90)
SCORE=$(node -e "console.log(JSON.parse(require('fs').readFileSync('lighthouse-results.json')).categories.accessibility.score * 100)")
echo "Lighthouse Accessibility Score: $SCORE"

if [ "$(echo "$SCORE < 90" | bc)" -eq 1 ]; then
  echo "Error: Accessibility score is below 90!"
  exit 1
fi

echo "Accessibility audit passed!"
