package handler

import (
	"context"

	"github.com/Housiadas/cerberus/internal/core/billing"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

func (h *Handler) CreateCheckoutSession(
	_ context.Context,
	_ openapi.CreateCheckoutSessionRequestObject,
) (openapi.CreateCheckoutSessionResponseObject, error) {
	return openapi.CreateCheckoutSession200JSONResponse(toOpenAPICheckoutResponse(billing.CheckoutResponse{})), nil
}

func (h *Handler) CreatePortalSession(
	_ context.Context,
	_ openapi.CreatePortalSessionRequestObject,
) (openapi.CreatePortalSessionResponseObject, error) {
	return openapi.CreatePortalSession200JSONResponse(toOpenAPIPortalResponse(billing.PortalResponse{})), nil
}

func (h *Handler) ListSubscriptions(
	_ context.Context,
	_ openapi.ListSubscriptionsRequestObject,
) (openapi.ListSubscriptionsResponseObject, error) {
	return openapi.ListSubscriptions200JSONResponse(toOpenAPISubscriptionResponses([]billing.SubscriptionResponse{})), nil
}

func (h *Handler) ListInvoices(
	_ context.Context,
	_ openapi.ListInvoicesRequestObject,
) (openapi.ListInvoicesResponseObject, error) {
	return openapi.ListInvoices200JSONResponse(toOpenAPIInvoiceResponses([]billing.InvoiceResponse{})), nil
}
