// Package billing_usecase maintains the use case layer for billing operations
package billing_usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	account2 "github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/internal/core/account/account_service"
	"github.com/Housiadas/cerberus/internal/core/invoice"
	"github.com/Housiadas/cerberus/internal/core/invoice/invoice_service"
	"github.com/Housiadas/cerberus/internal/core/subscription"
	"github.com/Housiadas/cerberus/internal/core/subscription/subscription_service"
	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/cursor"
	stripepkg "github.com/Housiadas/cerberus/pkg/stripe"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
)

// UseCase manages the set of APIs for billing use cases.
type UseCase struct {
	stripeClient    *stripepkg.Client
	accountSvc      *account_service.Service
	subscriptionSvc *subscription_service.Service
	invoiceSvc      *invoice_service.Service
	uuidGen         uuidgen.Generator
	clock           clock.Clock
}

// NewUseCase constructs a use case for billing operations.
func NewUseCase(
	stripeClient *stripepkg.Client,
	accountSvc *account_service.Service,
	subscriptionSvc *subscription_service.Service,
	invoiceSvc *invoice_service.Service,
	uuidGen uuidgen.Generator,
	clock clock.Clock,
) *UseCase {
	return &UseCase{
		stripeClient:    stripeClient,
		accountSvc:      accountSvc,
		subscriptionSvc: subscriptionSvc,
		invoiceSvc:      invoiceSvc,
		uuidGen:         uuidGen,
		clock:           clock,
	}
}

// CreateAccountWithStripe creates an account and a corresponding Stripe customer.
func (uc *UseCase) CreateAccountWithStripe(
	ctx context.Context,
	accountName, email string,
) (account2.Account, error) {
	cust, err := uc.stripeClient.CreateCustomer(ctx, accountName, email)
	if err != nil {
		return account2.Account{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"stripe create customer: %s",
			err,
		)
	}

	na := account2.NewAccount{Name: accountName}

	created, err := uc.accountSvc.Create(ctx, na)
	if err != nil {
		return account2.Account{}, fmt.Errorf("create account: %w", err)
	}

	// Update with Stripe customer ID
	ua := account2.UpdateAccount{
		StripeCustomerID: &sql.NullString{String: cust.ID, Valid: true},
	}

	updated, err := uc.accountSvc.Update(ctx, created, ua)
	if err != nil {
		return account2.Account{}, fmt.Errorf("update account stripe id: %w", err)
	}

	return updated, nil
}

// CreateCheckoutSession creates a Stripe Checkout session for an account.
func (uc *UseCase) CreateCheckoutSession(
	ctx context.Context,
	req CheckoutRequest,
) (CheckoutResponse, error) {
	accountUUID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return CheckoutResponse{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"invalid account id: %s",
			err,
		)
	}

	acc, err := uc.accountSvc.QueryByID(ctx, accountUUID)
	if err != nil {
		return CheckoutResponse{}, fmt.Errorf("query account: %w", err)
	}

	if !acc.StripeCustomerID().Valid {
		return CheckoutResponse{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"account has no stripe customer",
		)
	}

	sess, err := uc.stripeClient.CreateCheckoutSession(
		ctx,
		acc.StripeCustomerID().String,
		req.PriceID,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		return CheckoutResponse{}, fmt.Errorf("create checkout session: %w", err)
	}

	return CheckoutResponse{URL: sess.URL}, nil
}

// CreatePortalSession creates a Stripe Customer Portal session for an account.
func (uc *UseCase) CreatePortalSession(
	ctx context.Context,
	req PortalRequest,
) (PortalResponse, error) {
	accountUUID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return PortalResponse{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"invalid account id: %s",
			err,
		)
	}

	acc, err := uc.accountSvc.QueryByID(ctx, accountUUID)
	if err != nil {
		return PortalResponse{}, fmt.Errorf("query account: %w", err)
	}

	if !acc.StripeCustomerID().Valid {
		return PortalResponse{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"account has no stripe customer",
		)
	}

	sess, err := uc.stripeClient.CreateCustomerPortalSession(
		ctx,
		acc.StripeCustomerID().String,
		req.ReturnURL,
	)
	if err != nil {
		return PortalResponse{}, fmt.Errorf("create portal session: %w", err)
	}

	return PortalResponse{URL: sess.URL}, nil
}

// GetSubscriptions returns subscriptions for an account.
func (uc *UseCase) GetSubscriptions(
	ctx context.Context,
	accountID string,
) ([]SubscriptionResponse, error) {
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"invalid account id: %s",
			err,
		)
	}

	subs, err := uc.subscriptionSvc.QueryByAccountID(ctx, accountUUID)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}

	return toSubscriptionResponses(subs), nil
}

// GetInvoices returns invoices for an account.
func (uc *UseCase) GetInvoices(ctx context.Context, accountID string) ([]InvoiceResponse, error) {
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"invalid account id: %s",
			err,
		)
	}

	invs, err := uc.invoiceSvc.QueryByAccountID(ctx, accountUUID)
	if err != nil {
		return nil, fmt.Errorf("query invoices: %w", err)
	}

	return toInvoiceResponses(invs), nil
}

// HandleWebhookSubscription handles subscription webhook events.
func (uc *UseCase) HandleWebhookSubscription(
	ctx context.Context,
	stripeSub *stripe.Subscription,
) error {
	existing, err := uc.subscriptionSvc.QueryByStripeID(ctx, stripeSub.ID)
	if err != nil {
		// New subscription - create it
		return uc.createSubscriptionFromStripe(ctx, stripeSub)
	}

	// Update existing
	existing = existing.WithStatus(string(stripeSub.Status)).
		WithCancelAtPeriodEnd(stripeSub.CancelAtPeriodEnd).
		WithUpdatedAt(uc.clock.Now())

	if len(stripeSub.Items.Data) > 0 {
		item := stripeSub.Items.Data[0]

		if item.CurrentPeriodStart > 0 {
			t := time.Unix(item.CurrentPeriodStart, 0).UTC()
			existing = existing.WithCurrentPeriodStart(&t)
		}

		if item.CurrentPeriodEnd > 0 {
			t := time.Unix(item.CurrentPeriodEnd, 0).UTC()
			existing = existing.WithCurrentPeriodEnd(&t)
		}
	}

	if stripeSub.CanceledAt > 0 {
		t := time.Unix(stripeSub.CanceledAt, 0).UTC()
		existing = existing.WithCanceledAt(&t)
	}

	updateErr := uc.subscriptionSvc.Update(ctx, existing)
	if updateErr != nil {
		return fmt.Errorf("update subscription: %w", updateErr)
	}

	return nil
}

// HandleWebhookInvoice handles invoice webhook events.
func (uc *UseCase) HandleWebhookInvoice(ctx context.Context, stripeInv *stripe.Invoice) error {
	existing, err := uc.invoiceSvc.QueryByStripeID(ctx, stripeInv.ID)
	if err != nil {
		// New invoice - create it
		return uc.createInvoiceFromStripe(ctx, stripeInv)
	}

	existing = existing.
		WithStatus(string(stripeInv.Status)).
		WithAmountPaid(stripeInv.AmountPaid).
		WithUpdatedAt(uc.clock.Now())

	updateErr := uc.invoiceSvc.Update(ctx, existing)
	if updateErr != nil {
		return fmt.Errorf("update invoice: %w", updateErr)
	}

	return nil
}

func (uc *UseCase) createSubscriptionFromStripe(
	ctx context.Context,
	stripeSub *stripe.Subscription,
) error {
	id, err := uc.uuidGen.Generate()
	if err != nil {
		return fmt.Errorf("uuid error: %w", err)
	}

	// Look up account by stripe customer ID
	accountID, err := uc.findAccountIDByStripeCustomer(ctx, stripeSub.Customer.ID)
	if err != nil {
		return fmt.Errorf("find account: %w", err)
	}

	priceID := ""
	if len(stripeSub.Items.Data) > 0 {
		priceID = stripeSub.Items.Data[0].Price.ID
	}

	now := uc.clock.Now()

	var periodStart, periodEnd, canceledAt *time.Time

	if len(stripeSub.Items.Data) > 0 {
		item := stripeSub.Items.Data[0]

		if item.CurrentPeriodStart > 0 {
			t := time.Unix(item.CurrentPeriodStart, 0).UTC()
			periodStart = &t
		}

		if item.CurrentPeriodEnd > 0 {
			t := time.Unix(item.CurrentPeriodEnd, 0).UTC()
			periodEnd = &t
		}
	}

	if stripeSub.CanceledAt > 0 {
		t := time.Unix(stripeSub.CanceledAt, 0).UTC()
		canceledAt = &t
	}

	sub := subscription.New(
		id,
		accountID,
		stripeSub.ID,
		stripeSub.Customer.ID,
		priceID,
		string(stripeSub.Status),
		periodStart,
		periodEnd,
		stripeSub.CancelAtPeriodEnd,
		canceledAt,
		now,
		now,
	)

	createErr := uc.subscriptionSvc.Create(ctx, sub)
	if createErr != nil {
		return fmt.Errorf("create subscription: %w", createErr)
	}

	return nil
}

func (uc *UseCase) createInvoiceFromStripe(ctx context.Context, stripeInv *stripe.Invoice) error {
	id, err := uc.uuidGen.Generate()
	if err != nil {
		return fmt.Errorf("uuid error: %w", err)
	}

	accountID, err := uc.findAccountIDByStripeCustomer(ctx, stripeInv.Customer.ID)
	if err != nil {
		return fmt.Errorf("find account: %w", err)
	}

	now := uc.clock.Now()

	var periodStart, periodEnd *time.Time

	if stripeInv.PeriodStart > 0 {
		t := time.Unix(stripeInv.PeriodStart, 0).UTC()
		periodStart = &t
	}

	if stripeInv.PeriodEnd > 0 {
		t := time.Unix(stripeInv.PeriodEnd, 0).UTC()
		periodEnd = &t
	}

	invoiceURL := ""
	if stripeInv.HostedInvoiceURL != "" {
		invoiceURL = stripeInv.HostedInvoiceURL
	}

	inv := invoice.New(
		id,
		accountID,
		stripeInv.ID,
		stripeInv.Customer.ID,
		string(stripeInv.Status),
		stripeInv.AmountDue,
		stripeInv.AmountPaid,
		string(stripeInv.Currency),
		invoiceURL,
		periodStart,
		periodEnd,
		now,
		now,
	)

	createErr := uc.invoiceSvc.Create(ctx, inv)
	if createErr != nil {
		return fmt.Errorf("create invoice: %w", createErr)
	}

	return nil
}

func (uc *UseCase) findAccountIDByStripeCustomer(
	ctx context.Context,
	stripeCustomerID string,
) (uuid.UUID, error) {
	filter := account2.QueryFilter{
		StripeCustomerID: &stripeCustomerID,
	}

	accs, err := uc.accountSvc.Query(ctx, filter, account2.GetDefaultOrderBy(), defaultCursor())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("query accounts: %w", err)
	}

	if len(accs) == 0 {
		return uuid.UUID{}, errs.Errorf(
			errs.NotFound,
			errs.CodeAccountNotFound,
			"no account found for stripe customer: %s",
			stripeCustomerID,
		)
	}

	return accs[0].ID(), nil
}

func defaultCursor() cursor.Cursor {
	c, _ := cursor.Parse("", "100")

	return c
}

func toSubscriptionResponses(subs []subscription.Subscription) []SubscriptionResponse {
	resp := make([]SubscriptionResponse, len(subs))
	for i, sub := range subs {
		resp[i] = SubscriptionResponse{
			ID:                   sub.ID().String(),
			StripeSubscriptionID: sub.StripeSubscriptionID(),
			StripePriceID:        sub.StripePriceID(),
			Status:               sub.Status(),
			CancelAtPeriodEnd:    sub.CancelAtPeriodEnd(),
			CreatedAt:            sub.CreatedAt().UTC().Format(time.RFC3339),
		}

		if sub.CurrentPeriodStart() != nil {
			s := sub.CurrentPeriodStart().UTC().Format(time.RFC3339)
			resp[i].CurrentPeriodStart = &s
		}

		if sub.CurrentPeriodEnd() != nil {
			s := sub.CurrentPeriodEnd().UTC().Format(time.RFC3339)
			resp[i].CurrentPeriodEnd = &s
		}
	}

	return resp
}

func toInvoiceResponses(invs []invoice.Invoice) []InvoiceResponse {
	resp := make([]InvoiceResponse, len(invs))
	for i, inv := range invs {
		resp[i] = InvoiceResponse{
			ID:              inv.ID().String(),
			StripeInvoiceID: inv.StripeInvoiceID(),
			Status:          inv.Status(),
			AmountDue:       inv.AmountDue(),
			AmountPaid:      inv.AmountPaid(),
			Currency:        inv.Currency(),
			InvoiceURL:      inv.InvoiceURL(),
			CreatedAt:       inv.CreatedAt().UTC().Format(time.RFC3339),
		}
	}

	return resp
}
