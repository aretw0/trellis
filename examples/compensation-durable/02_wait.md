---
options:
  - text: "approve"
    to: 03_flight
  - text: "reject"
    to: 03_rejection
on_signal:
  manager_approval: 03_flight
  manager_rejection: 03_rejection
---

Waiting for Manager Approval... (Press Ctrl+C to simulate Time Passing, or type 'approve'/'reject' now)
