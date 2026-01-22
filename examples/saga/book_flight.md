---
type: tool
tool_call:
  name: book_flight
metadata:
  undo_action: cancel_flight
save_to: flight_id
on_error: rollback
transitions:
  - to: book_hotel
---
Booking Flight...
