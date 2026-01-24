// 1. Input: Read from Environment Variables
// Trellis capitalizes argument names: 'name' -> 'TRELLIS_ARG_NAME'
const name = process.env.TRELLIS_ARG_NAME || "Guest";
const greeting = process.env.TRELLIS_ARG_GREETING || "Hi";

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
