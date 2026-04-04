package handler

import (
	"context"
	"errors"

	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

var errBillingNotConfigured = errors.New("billing service not configured")

func (h *Handler) StripeWebhook(
	_ context.Context,
	_ openapi.StripeWebhookRequestObject,
) (openapi.StripeWebhookResponseObject, error) {
	// Note: In production, Stripe webhooks require raw body + Stripe-Signature header
	// for verification via ConstructWebhookEvent. The strict server interface
	// parses the body as JSON, so a custom chi route should be used instead
	//  of proper signature verification. This is a placeholder implementation.
	if h.svc.billing == nil {
		return nil, errBillingNotConfigured
	}

	return openapi.StripeWebhook200Response{}, nil
}
