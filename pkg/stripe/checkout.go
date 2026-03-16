package stripe

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v82"
)

// CreateCheckoutSession creates a Stripe Checkout session.
func (c *Client) CreateCheckoutSession(
	ctx context.Context,
	customerID, priceID, successURL, cancelURL string,
) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionCreateParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}

	sess, err := c.client.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe create checkout session: %w", err)
	}

	return sess, nil
}
