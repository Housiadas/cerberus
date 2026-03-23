// Package stripe provides a wrapper around the Stripe Go SDK.
package stripe

import (
	"encoding/json"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/stripe/stripe-go/v82"
)

type Config struct {
	WebhookSecret string
	SecretKey     string
	Log           logger.Logger
}

// Client wraps the Stripe SDK for billing operations.
type Client struct {
	webhookSecret string
	log           logger.Logger
	client        *stripe.Client
}

// New creates a new Stripe client. It sets the global API key.
func New(cfg Config) *Client {
	cstripe := stripe.NewClient(cfg.SecretKey)

	return &Client{
		log:           cfg.Log,
		client:        cstripe,
		webhookSecret: cfg.WebhookSecret,
	}
}

// EventDataTo unmarshals the event data object into the provided target.
func EventDataTo(event stripe.Event, target any) error {
	err := json.Unmarshal(event.Data.Raw, target)
	if err != nil {
		return fmt.Errorf("stripe event unmarshal: %w", err)
	}

	return nil
}
