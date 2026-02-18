---
id: loop
default_context:
  count: 0
do:
  name: increment
  args:
    count: "{{ .count }}"
save_to: count
to: loop-result
on_denied: menu
---
# Loop Iteration #{{ .count }}

Moving back to menu
