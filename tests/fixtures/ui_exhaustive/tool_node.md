---
do:
  name: mock_tool
  args:
    input: "{{ .user_input }}"
save_to: "tool_result"
on_error: error_node
to: tool_result_node
---
Executing tool step.
