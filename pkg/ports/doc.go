/*
Package ports defines the driven ports (interfaces) for the Trellis engine.

These interfaces decouple the core logic from external implementations, allowing
the engine to work with various storage backends, graph sources, and signal managers.

# Key Interfaces

  - GraphLoader: Responsible for loading Node definitions (e.g., from Loam or Memory).
  - StateStore: Responsible for persisting and loading session State.
  - DistributedLocker: Provides distributed locking for handling concurrent session access.
*/
package ports
