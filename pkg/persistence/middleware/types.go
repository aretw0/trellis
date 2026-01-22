package middleware

import "github.com/aretw0/trellis/pkg/ports"

// Middleware allows wrapping a StateStore to add behavior.
type Middleware func(ports.StateStore) ports.StateStore
