// Package stripe provides a wrapper around the Stripe Go SDK.
package stripe

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v82"
)

type logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Debugc(ctx context.Context, caller int, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

type Config struct {
	WebhookSecret string
	SecretKey     string
	Log           logger
}

// Client wraps the Stripe SDK for billing operations.
type Client struct {
	webhookSecret string
	log           logger
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
