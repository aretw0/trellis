---
type: input
transitions:
  - to: finish
    condition: input == 'next'
on_signal:
  interrupt: finish
---

# Welcome to Reactivity Demo

This is the start node.
Enter 'next' to proceed.
