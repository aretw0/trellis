---
type: tool
do:
  name: reserve_hotel
  args:
    city: "Paris"

undo:
  name: cancel_hotel
  args:
    id: "{{ .hotel.id }}"

save_to: hotel
to: 02_wait
---
Reserving Hotel in Paris...
