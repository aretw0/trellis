---
do:
  name: book_flight
metadata:
  undo_action: cancel_flight
save_to: flight_id
on_error: manual_rollback
to: book_hotel
---
Booking Flight...
