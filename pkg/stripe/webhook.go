package stripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

// ConstructWebhookEvent verifies and parses a Stripe webhook event.
func (c *Client) ConstructWebhookEvent(payload []byte, signature string) (stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, c.webhookSecret)
	if err != nil {
		return stripe.Event{}, fmt.Errorf("stripe webhook verification: %w", err)
	}

	return event, nil
}
