/*
Package runner implements the execution loop and I/O orchestration for the Trellis engine.

It acts as the bridge between the core state machine (Engine) and the outside world.
The runner manages session persistence, handles user input/output through pluggable
handlers, and integrates with the 'lifecycle' library for signal handling and supervision.

# Key Components

  - Runner: The main orchestrator that implements the lifecycle.Worker interface.
  - InputHandler: Decouples how the engine receives inputs (CLI, JSON, etc.).
  - TextHandler: A standard implementation for interactive CLI usage.

# Usage

	r := runner.NewRunner(
		runner.WithEngine(engine),
		runner.WithSessionID("user-1"),
		runner.WithInputHandler(runner.NewTextHandler(os.Stdout, runner.WithStdin())),
	)

	if err := r.Run(ctx); err != nil {
		log.Fatal(err)
	}
*/
package runner
