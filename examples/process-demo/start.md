start:
  type: text
  content: "Starting Process Demo..."
  transitions:
    - to: call_script

call_script:
  type: tool
  tool_call:
    name: hello_script
    args:
      name: "TrellisUser"
  save_to: script_output
  transitions:
    - to: show_result

show_result:
  type: text
  content: |
    Script Output:
    ```json
    {{ .script_output }}
    ```
  transitions:
    - to: end

end:
  type: text
  content: "Demo Complete."
