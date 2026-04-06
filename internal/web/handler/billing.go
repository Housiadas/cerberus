package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/billing"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

func (h *Handler) CreateCheckoutSession(
	ctx context.Context,
	request openapi.CreateCheckoutSessionRequestObject,
) (openapi.CreateCheckoutSessionResponseObject, error) {
	resp, err := h.svc.billing.CreateCheckoutSession(ctx, billing.CheckoutRequest{
		AccountID:  request.Body.AccountId,
		PriceID:    request.Body.PriceId,
		SuccessURL: request.Body.SuccessUrl,
		CancelURL:  request.Body.CancelUrl,
	})
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}

	return openapi.CreateCheckoutSession200JSONResponse(toOpenAPICheckoutResponse(resp)), nil
}

func (h *Handler) CreatePortalSession(
	ctx context.Context,
	request openapi.CreatePortalSessionRequestObject,
) (openapi.CreatePortalSessionResponseObject, error) {
	resp, err := h.svc.billing.CreatePortalSession(ctx, billing.PortalRequest{
		AccountID: request.Body.AccountId,
		ReturnURL: request.Body.ReturnUrl,
	})
	if err != nil {
		return nil, fmt.Errorf("create portal session: %w", err)
	}

	return openapi.CreatePortalSession200JSONResponse(toOpenAPIPortalResponse(resp)), nil
}

func (h *Handler) ListSubscriptions(
	ctx context.Context,
	request openapi.ListSubscriptionsRequestObject,
) (openapi.ListSubscriptionsResponseObject, error) {
	subs, err := h.svc.billing.GetSubscriptions(ctx, request.Params.AccountId)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	return openapi.ListSubscriptions200JSONResponse(toOpenAPISubscriptionResponses(subs)), nil
}

func (h *Handler) ListInvoices(
	ctx context.Context,
	request openapi.ListInvoicesRequestObject,
) (openapi.ListInvoicesResponseObject, error) {
	invs, err := h.svc.billing.GetInvoices(ctx, request.Params.AccountId)
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}

	return openapi.ListInvoices200JSONResponse(toOpenAPIInvoiceResponses(invs)), nil
}
