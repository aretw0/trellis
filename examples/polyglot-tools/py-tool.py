import os
import json
import sys

def main():
    # 1. Input: Read from Environment Variables
    # Trellis passes arguments prefixed with TRELLIS_ARG_
    name = os.getenv("TRELLIS_ARG_NAME", "Guest")
    greeting = os.getenv("TRELLIS_ARG_GREETING", "Hello")

    # [Demo] Handling Complex Arguments (JSON)
    # If the flow passes a map/object, Trellis serializes it as a JSON string.
    raw_config = os.getenv("TRELLIS_ARG_CONFIG", "{}")
    try:
        config = json.loads(raw_config)
    except json.JSONDecodeError:
        config = {"error": "Invalid JSON in CONFIG", "raw": raw_config}

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
