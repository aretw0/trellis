---
type: question
input_type: choice
input_options: ["Go to End", "Go Loop"]
input_default: "Go to End"
transitions:
  - to: success
    condition: input == "Go to End"
  - to: start
    condition: input == "Go Loop"
---
# Interactive Inputs

This is a **Question Node**. The Engine pauses here and asks the Host (CLI) to collect input.

We specified:

- `input_type`: `choice`
- `input_options`: `["Go to End", "Go Loop"]`

Select an option below:
