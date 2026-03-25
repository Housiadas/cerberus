package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

func (h *Handler) CreateCheckoutSession(
	ctx context.Context,
	request openapi.CreateCheckoutSessionRequestObject,
) (openapi.CreateCheckoutSessionResponseObject, error) {
	resp, err := h.usecase.billing.CreateCheckoutSession(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}

	return openapi.CreateCheckoutSession200JSONResponse(resp), nil
}

func (h *Handler) CreatePortalSession(
	ctx context.Context,
	request openapi.CreatePortalSessionRequestObject,
) (openapi.CreatePortalSessionResponseObject, error) {
	resp, err := h.usecase.billing.CreatePortalSession(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("create portal session: %w", err)
	}

	return openapi.CreatePortalSession200JSONResponse(resp), nil
}

func (h *Handler) ListSubscriptions(
	ctx context.Context,
	request openapi.ListSubscriptionsRequestObject,
) (openapi.ListSubscriptionsResponseObject, error) {
	subs, err := h.usecase.billing.GetSubscriptions(ctx, request.Params.AccountId)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	return openapi.ListSubscriptions200JSONResponse(subs), nil
}

func (h *Handler) ListInvoices(
	ctx context.Context,
	request openapi.ListInvoicesRequestObject,
) (openapi.ListInvoicesResponseObject, error) {
	invs, err := h.usecase.billing.GetInvoices(ctx, request.Params.AccountId)
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}

	return openapi.ListInvoices200JSONResponse(invs), nil
}
