package eventbus

import (
	"context"

	"github.com/Housiadas/cerberus/internal/types/event"
)

// NopDispatcher is a no-op event dispatcher for use in tests
// and CLI commands where event dispatching is not needed.
type NopDispatcher struct{}

// NewNop creates a NopDispatcher.
func NewNop() *NopDispatcher {
	return &NopDispatcher{}
}

// Dispatch is a no-op that always returns nil.
func (d *NopDispatcher) Dispatch(
	_ context.Context,
	_ event.DomainEvent,
) error {
	return nil
}
