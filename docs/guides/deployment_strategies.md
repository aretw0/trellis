# Deployment Strategies & Resource Management

Trellis is designed to be flexible, running either as an embedded library inside your Go application or as a standalone service (Sidecar/MCP). Choosing the right strategy depends on your isolation requirements and traffic patterns.

## 1. Embedded (Library)

In this mode, Trellis runs inside your Host process. This offers the lowest latency and simplest deployment but shares memory resources with your application.

### Memory Provisioning Formula

Since Go processes share a single Heap, you must account for Trellis's memory footprint to prevent OOM (Out of Memory) kills.

**Estimated RAM Usage:**

```text
Total RAM = App Base + (Concurrent Sessions * Session Overhead)
```

Where:

* **App Base**: Your application's baseline memory.
* **Session Overhead**: Fixed Context (~1KB) + **Input Buffer**.

### The Input Buffer Risk

If a user sends a large input (e.g., 10MB text), that memory is allocated in your process. Multiplied by 1000 users, this causes **Memory Starvation**.

**Mitigation: Input Sanitization**
Trellis enforces a strict input limit to make this variable predictable.

* **Default Limit**: 4KB (`4096 bytes`).
* **Configuration**: Set `TRELLIS_MAX_INPUT_SIZE` env var.

**Example**:
If you set `TRELLIS_MAX_INPUT_SIZE=65536` (64KB) and expect 100 concurrent sessions:
`RAM Spike Risk = 100 * 64KB = 6.4MB`. This is safe and predictable.

## 2. Sidecar (Standalone Service)

For high-traffic or multi-tenant environments (e.g., Kubernetes Pods), you may want to isolate Trellis to protect your main application from resource contention.

### Architecture

Run Trellis as a separate container or process using `trellis serve`:

* **Container A (Host App)**: Your business logic.
* **Container B (Trellis)**: The cognitive engine.

### Benefits

* **OS-Level Isolation**: You can set strict k8s resource limits (e.g., `resources: limits: memory: 256Mi`) for the Trellis container.
* **Crash Safety**: If Trellis crashes or OOMs due to a complex flow, your main app stays alive.
* **Independent Scaling**: Scale the Trellis tier independently of your backend.

## Observability & Security

Trellis emits structured logs (`slog` compatible) for security events. Monitor these to detect attacks or misconfiguration.

### Security Events

* **Input Rejected**: Emitted when a user exceeds the size limit.
  * Level: `WARN`
  * Message: `Input Rejected`
  * Attributes: `reason`, `size`, `limit`.

**Example Log:**

```json
{"time":"...","level":"WARN","msg":"Input Rejected","size":10500,"limit":4096,"reason":"input exceeds maximum allowed size"}
```

**Alerting Strategy**:
Configure your APM (Prometheus/Datadog) to alert if `Input Rejected` counts spike, as this may indicate a DoS attempt or a legitimate need to increase `TRELLIS_MAX_INPUT_SIZE`.
