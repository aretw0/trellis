---
id: start
type: question
input_type: choice
input_options: ["A", "B"]
transitions:
  - to: end_a
    condition: input == "A"
  - to: end_b
    condition: input == "B"
---
Start Node
