---
do:
  name: book_hotel
metadata:
  undo_action: cancel_hotel
save_to: hotel_id
on_error: manual_rollback
to: book_car
---
Booking Hotel...
