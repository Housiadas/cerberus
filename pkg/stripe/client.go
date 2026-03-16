// Package stripe provides a wrapper around the Stripe Go SDK.
package stripe

import (
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v82"
	billingportalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

// Client wraps the Stripe SDK for billing operations.
type Client struct {
	webhookSecret string
}

// New creates a new Stripe client. It sets the global API key.
func New(secretKey, webhookSecret string) *Client {
	stripe.Key = secretKey

	return &Client{
		webhookSecret: webhookSecret,
	}
}

// CreateCustomer creates a Stripe customer.
func (c *Client) CreateCustomer(name, email string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Name:  stripe.String(name),
		Email: stripe.String(email),
	}

	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe create customer: %w", err)
	}

	return cust, nil
}

// CreateSubscription creates a Stripe subscription for a customer.
func (c *Client) CreateSubscription(customerID, priceID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe create subscription: %w", err)
	}

	return sub, nil
}

// CancelSubscription cancels a Stripe subscription.
func (c *Client) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionCancelParams{}

	sub, err := subscription.Cancel(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("stripe cancel subscription: %w", err)
	}

	return sub, nil
}

// CreateCheckoutSession creates a Stripe Checkout session.
func (c *Client) CreateCheckoutSession(
	customerID, priceID, successURL, cancelURL string,
) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}

	sess, err := checkoutsession.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe create checkout session: %w", err)
	}

	return sess, nil
}

// CreateCustomerPortalSession creates a Stripe Customer Portal session.
func (c *Client) CreateCustomerPortalSession(
	customerID, returnURL string,
) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := billingportalsession.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe create portal session: %w", err)
	}

	return sess, nil
}

// ConstructWebhookEvent verifies and parses a Stripe webhook event.
func (c *Client) ConstructWebhookEvent(payload []byte, signature string) (stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, c.webhookSecret)
	if err != nil {
		return stripe.Event{}, fmt.Errorf("stripe webhook verification: %w", err)
	}

	return event, nil
}

// EventDataTo unmarshals the event data object into the provided target.
func EventDataTo(event stripe.Event, target any) error {
	err := json.Unmarshal(event.Data.Raw, target)
	if err != nil {
		return fmt.Errorf("stripe event unmarshal: %w", err)
	}

	return nil
}
