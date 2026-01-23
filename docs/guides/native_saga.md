# Native SAGA Orchestration

The **SAGA Pattern** is a failure management pattern for distributed applications. It manages long-running transactions by coordinating a sequence of local transactions, each with a corresponding **Compensating Transaction** (Undo) to reverse the work if a future step fails.

Trellis v0.7 implements SAGA natively, meaning you don't need to manually wire "error paths" or complex conditionals in your graph.

## How it Works

The mechanics rely on three keywords in your Node definition:

1. **`do`**: The primary action (tool call) to execute.
2. **`undo`**: The action to execute if the Engine needs to rollback.
3. **`on_error: rollback`**: A special transition instruction that triggers the Rollback lifecycle.

### Example: Booking a Trip

Imagine a flow where you need to:

1. Reserve a Hotel.
2. Book a Flight.
3. Send an Email.

If "Book Flight" fails, we must cancel the "Reserve Hotel" action to avoid charging the user for a hotel they can't reach.

#### 1. Reserve Hotel (Step 1)

```yaml
# 01_hotel.md
type: tool
do:
  name: reserve_hotel
  args:
    city: "Paris"
    date: "2024-12-25"

# Define the Undo action right here
undo:
  name: cancel_hotel
  args:
    reservation_id: "{{ .hotel_reservation_id }}" # Captured from result

save_to: hotel_result # Data needed for next steps AND for undo arguments
to: 02_flight
```

#### 2. Book Flight (Step 2 - The Failure Point)

```yaml
# 02_flight.md
type: tool
do:
  name: book_flight
  args:
    destination: "CDG"

# If this fails, Trigger Rollback!
on_error: rollback
```

## The Rollback Process

When `on_error: rollback` is triggered:

1. The Engine changes status to `RollingBack`.
2. It looks at the **History Stack** (e.g., `["start", "01_hotel", "02_flight"]`).
3. It pops the current failed node (`02_flight`). No undo needed as it failed.
4. It pops the previous node (`01_hotel`).
5. It detects the `undo` definition in `01_hotel`.
6. It executes the `cancel_hotel` tool.
7. It continues unwinding until the stack is empty or it hits a Savepoint (Start).

## Best Practices

### Locality of Behavior (LoB)

Always define the `undo` logic in the **same file** as the `do` logic. This ensures that developers modifying the primary action are reminded to update the compensating transaction.

### Idempotency

Your Undo tools should be **Idempotent**. The Engine guarantees `at-most-once` execution for the primary action, but during a crash recovery scenario, it might attempt to rollback again. Your tools should handle "cancelling an already cancelled reservation" gracefully (e.g., return success).

### Variables in Undo

The `undo` action has access to the **Full Session Context** at the time of the rollback. This includes data saved by the `do` action (via `save_to`).

For example, if `reserve_hotel` saves `{ "id": "123" }` to `hotel_result`, your undo args can reference `{{ .hotel_result.id }}`.

## Ready to Run?

Check out the complete working example in [`examples/compensation-native`](../../examples/compensation-native).

```bash
go run ./examples/compensation-native
```
