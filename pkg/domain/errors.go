package domain

import "errors"

// ErrUnhandledSignal is returned when a signal is received but no handler is defined for it.
var ErrUnhandledSignal = errors.New("unhandled signal")

// ErrSessionNotFound is returned when a session ID cannot be found in the store.
var ErrSessionNotFound = errors.New("session not found")
