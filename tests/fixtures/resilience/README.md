# Resilience Fixtures

This directory contains programs that simulate various "behaviors" of external processes. They are used to ensure the Trellis Process Adapter can handle any situation.

## Fixtures

### 1. `good_citizen`
A "well-behaved" process.

- **Behavior:** Starts, waits for a signal (SIGINT/SIGTERM), and exits with code 0 immediately upon receiving it.
- **Test Case:** Verifies graceful shutdown logic.

### 2. `bad_citizen_ignore`
A process that refuses to die.

- **Behavior:** Starts, catches signals, but prints a message and *continues running* instead of exiting.
- **Test Case:** Verifies that Trellis enforces timeouts and performs a "Force Kill" when a process is unresponsive.

### 3. `bad_citizen_slow`
A process that is slow to shutdown.

- **Behavior:** Starts, catches a signal, sleeps for a long time, then exits.
- **Test Case:** Verifies wait logic and grace periods.

### 4. `crashy`
A process that fails immediately.

- **Behavior:** Panics or exits with a non-zero code immediately on startup.
- **Test Case:** Verifies error reporting and crash detection.
