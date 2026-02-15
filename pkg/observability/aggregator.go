package observability

import (
	"context"

	"github.com/aretw0/introspection"
)

// Aggregator combines multiple watchers into a single view.
// It wraps introspection.AggregateWatchers to provide a simple wrapper.
type Aggregator struct {
	watchers []interface{}
}

// NewAggregator creates a new aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{
		watchers: make([]interface{}, 0),
	}
}

// AddWatcher registers a watcher (must implement Watch method).
func (a *Aggregator) AddWatcher(w interface{}) {
	a.watchers = append(a.watchers, w)
}

// Watch returns the aggregated state channel.
func (a *Aggregator) Watch(ctx context.Context) <-chan introspection.StateSnapshot {
	return introspection.AggregateWatchers(ctx, a.watchers...)
}
