import os
import json
import sys

def main():
    # 1. Input: Read from TRELLIS_ARGS (2026-02-17: Tool Argument Evolution)
    # Trellis passes all arguments as a JSON object in TRELLIS_ARGS
    raw_args = os.getenv("TRELLIS_ARGS", "{}")
    try:
        args = json.loads(raw_args)
    except json.JSONDecodeError:
        args = {}

    name = args.get("name", "Guest")
    greeting = args.get("greeting", "Hello")
    config = args.get("config", {})

    # 2. Logic: Perform some operation
    message = f"{greeting}, {name}! [Python]"
    
    # Use config data if present (e.g. "debug": true)
    if config.get("debug"):
        message += f" (Debug Mode: {config})"

    # 3. Output: Return result as JSON to Stdout
    output = {
        "message": message,
        "runtime": f"Python {sys.version.split()[0]}",
        "config_received": config,
        "status": "success"
    }

    # Print JSON to stdout
    print(json.dumps(output))

if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        # Error Handling: Print to stderr
        print(f"Error in python script: {e}", file=sys.stderr)
        sys.exit(1)
