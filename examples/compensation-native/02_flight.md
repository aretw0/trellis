---
do:
  name: book_flight
  args:
    destination: "CDG"

undo:
  name: cancel_flight
  args:
    id: "{{ .flight.id }}"

save_to: flight
to: 03_car
---
Booking Flight to CDG...
