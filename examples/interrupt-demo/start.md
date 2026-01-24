---
wait: true
on_signal:
  interrupt: confirm_exit
---

# Interrupt Demo

This flow demonstrates handling Ctrl+C (Interrupt Signal).

Press **Ctrl+C** now to verify the behavior.
If handled correctly, it should ask for confirmation.
If not, the program will exit immediately.

Or press Enter to finish normally.
