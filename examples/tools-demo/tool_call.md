---
id: tool_call
type: tool
tool_call:
  name: echo
  args:
    msg: "Hello from Tool"
metadata:
  confirm_msg: "DANGER: You are about to echo a message. Proceed? [y/N]"
save_to: echo_output
transitions:
  - to: success
---
