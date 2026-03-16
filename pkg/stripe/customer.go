package stripe

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v82"
)

// CreateCustomer creates a Stripe customer.
func (c *Client) CreateCustomer(ctx context.Context, name, email string) (*stripe.Customer, error) {
	params := &stripe.CustomerCreateParams{
		Name:  stripe.String(name),
		Email: stripe.String(email),
	}

	cust, err := c.client.V1Customers.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe create customer: %w", err)
	}

	return cust, nil
}
