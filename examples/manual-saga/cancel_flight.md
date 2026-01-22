---
type: tool
tool_call:
  name: cancel_flight
on_error: end_rollback
transitions:
  - to: end_rollback
---
Canceling Flight Reservation...
