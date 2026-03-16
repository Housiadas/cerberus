package stripe

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v82"
)

// CreateCustomerPortalSession creates a Stripe Customer Portal session.
func (c *Client) CreateCustomerPortalSession(
	ctx context.Context,
	customerID string,
	returnURL string,
) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := c.client.V1BillingPortalSessions.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe create portal session: %w", err)
	}

	return sess, nil
}
