package runner

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// SignalManager handles the complexity of OS signals, context cancellation,
// and platform-specific race conditions (e.g. Windows Stdin EOF vs Interrupt).
type SignalManager struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSignalManager creates a new manager and immediately starts listening for signals.
func NewSignalManager() *SignalManager {
	sm := &SignalManager{}
	sm.Reset()
	return sm
}

// Context returns the current signal context.
func (sm *SignalManager) Context() context.Context {
	return sm.ctx
}

// Reset re-arms the signal listener.
// Should be called after a signal has been successfully handled/intercepted
// to allow capturing subsequent signals.
func (sm *SignalManager) Reset() {
	if sm.cancel != nil {
		sm.cancel()
	}
	// We capture SIGINT (Ctrl+C) and SIGTERM
	sm.ctx, sm.cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// Stop permanently stops the signal listener.
func (sm *SignalManager) Stop() {
	if sm.cancel != nil {
		sm.cancel()
	}
}

// CheckRace waits briefly to see if a context cancellation follows an error.
// This is specifically to mitigate a race condition on Windows/PowerShell where
// Ctrl+C causes an EOF or Input Error slightly before the signal context is cancelled.
func (sm *SignalManager) CheckRace() {
	if sm.ctx.Err() == nil {
		select {
		case <-sm.ctx.Done():
			// Signal arrived during wait
		case <-time.After(100 * time.Millisecond):
			// Timeout, likely genuine error
		}
	}
}
