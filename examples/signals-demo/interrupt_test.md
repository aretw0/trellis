---
wait: true
on_signal:
  interrupt: signal_received
---
# Interrupt Test

I am waiting here indefinitely (until you press Enter or send a signal).

To trigger a signal, open another terminal and run:

```bash
curl -X POST http://localhost:8080/signal -H "Content-Type: application/json" -d '{"state": {}, "signal": "interrupt"}'
```

(Note: You must be running this flow with the HTTP server enabled for the API to work, but the Engine supports signals regardless of transport.)
