---
type: text
to: 01_hotel
---
# 🌍 Native SAGA Demo

This example demonstrates Trellis **Native SAGA orchestration**.

We will attempt a travel booking transaction:

1. 🏨 Reserve Hotel (Reversible)
2. ✈️ Book Flight (Reversible)
3. 🚗 Rent Car (Will Fail)

Watch the logs as the Engine automatically rolls back the first two steps when the third fails.
