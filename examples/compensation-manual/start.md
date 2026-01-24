---
save_to: booking_request
to: book_flight
---

# ✈️ Travel Booker

Starting automated booking sequence (SAGA Pattern).

This flow will attempt to:

- Book a Flight
- Book a Hotel
- Book a Car (Simulated Failure)

Upon failure, the system will automatically rollback the previous steps.

Press Enter to begin.
