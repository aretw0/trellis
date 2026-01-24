---
id: start
type: tool
do:
  name: get_os
  metadata:
    x-exec:
      command: go
      args: ["env", "GOOS"]
save_to: os_name
transitions:
  - to: end
---
# Inline Tool Demo

This node executes a command defined entirely in its own metadata (no tools.yaml required).
Running `go env GOOS`...
