/*
Package domain contains the core domain models and business logic for the Trellis engine.

It defines the fundamental entities of the state machine, such as Nodes, Transitions,
and the Execution State. This package is kept pure and free of external dependencies
like I/O or persistence, following Hexagonal Architecture principles.

# Key Entities

  - Node: Represents a point in the graph (Text, Input, or Tool/Action).
  - Transition: Defines the rules for moving from one node to another.
  - State: Captures the runtime snapshot of a session (Current Node, Context, History).
  - ActionRequest: A structural representation of what the host should render or execute.
*/
package domain
