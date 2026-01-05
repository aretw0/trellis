/*
Package trellis is a state machine engine designed for building text-based adventure games, interactive stories, and complex conversational flows.

It provides a flexible runtime that separates the narrative graph definition from the execution state, enabling rich, logic-driven navigation.

# Key Features

  - Graph-based State Machine: Define complex flows with nodes and transitions.
  - Pluggable Loaders: Load graphs from the filesystem (Markdown/Frontmatter) or in-memory structures.
  - Conditional Logic: Dynamic transitions based on custom evaluators.
  - State Management: Serializable state for long-running sessions.

# Basic Usage

Initialize the engine seamlessly using the filesystem loader (powered by Loam):

	eng, err := trellis.New("./story-data")
	if err != nil {
	    log.Fatal(err)
	}

	state := eng.Start()

	// Render the initial view
	actions, _, err := eng.Render(context.Background(), state)

	// Navigate based on input
	nextState, err := eng.Navigate(context.Background(), state, "open door")

For advanced usage, including custom loaders or conditional logic, refer to the examples and sub-packages.
*/
package trellis
