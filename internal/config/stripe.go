package config

// Stripe holds the Stripe API configuration.
type Stripe struct {
	SecretKey     string
	WebhookSecret string
}
