---
do:
  name: rent_car
  args:
    type: "compact"

# ⚠️ Trigger Native Rollback on Error
on_error: rollback
---
Attempting to rent a car...
