package stripe

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v82"
)

// CreateSubscription creates a Stripe subscription for a customer.
func (c *Client) CreateSubscription(
	ctx context.Context,
	customerID, priceID string,
) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionCreateParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionCreateItemParams{
			{
				Price: stripe.String(priceID),
			},
		},
	}

	sub, err := c.client.V1Subscriptions.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe create subscription: %w", err)
	}

	return sub, nil
}

// CancelSubscription cancels a Stripe subscription.
func (c *Client) CancelSubscription(
	ctx context.Context,
	subscriptionID string,
) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionCancelParams{}

	sub, err := c.client.V1Subscriptions.Cancel(ctx, subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("stripe cancel subscription: %w", err)
	}

	return sub, nil
}
