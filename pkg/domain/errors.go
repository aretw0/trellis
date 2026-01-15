package domain

import "errors"

// ErrUnhandledSignal is returned when a signal is received but no handler is defined for it.
var ErrUnhandledSignal = errors.New("unhandled signal")
