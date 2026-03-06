---
do:
  name: mock_tool
  args:
    input: "{{ .user_input }}"
on_error: error_node
to: tool_result_node
---
Executing tool step.
