package ports

import "github.com/aretw0/trellis/pkg/domain"

// ActionDispatcher defines how side-effects are executed.
// The engine emits requests, and the host implements this interface to handle them.
type ActionDispatcher interface {
	Dispatch(req domain.ActionRequest) (domain.ActionResponse, error)
}
