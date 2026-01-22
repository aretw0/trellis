---
type: start
save_to: booking_request
transitions:
- to: book_flight
---

# ✈️ Travel Booker

Starting recursive booking flow (SAGA Pattern).

Next steps:

1. Book Flight
2. Book Hotel
3. Book Car

(Note: Car booking is simulated to FAIL to trigger compensation).
