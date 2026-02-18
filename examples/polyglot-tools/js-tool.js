// 1. Input: Read from TRELLIS_ARGS (2026-02-17: Tool Argument Evolution)
// Trellis passes all arguments as a JSON object in TRELLIS_ARGS
const rawArgs = process.env.TRELLIS_ARGS || "{}";
const args = JSON.parse(rawArgs);
const name = args.name || "Guest";
const greeting = args.greeting || "Hi";

try {
  // 2. Logic
  const message = `${greeting}, ${name}! [Node.js]`;

  // 3. Output: JSON to Stdout
  const output = {
    message: message,
    runtime: `Node ${process.version}`,
    status: "success",
  };

  console.log(JSON.stringify(output));
} catch (error) {
  // Error: Print to Stderr
  console.error("Error in node script:", error);
  process.exit(1);
}
