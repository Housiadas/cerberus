// Package billing maintains the service layer for billing operations.
package billing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/internal/core/invoice"
	"github.com/Housiadas/cerberus/internal/core/subscription"
	errs "github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/pkg/cursor"
	stripepkg "github.com/Housiadas/cerberus/pkg/stripe"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// Service manages the set of APIs for billing operations.
type Service struct {
	stripeClient    *stripepkg.Client
	accountSvc      *account.Service
	subscriptionSvc *subscription.Service
	invoiceSvc      *invoice.Service
	uuidGen         generator
	clock           clock
}

// NewService constructs a service for billing operations.
func NewService(
	stripeClient *stripepkg.Client,
	accountSvc *account.Service,
	subscriptionSvc *subscription.Service,
	invoiceSvc *invoice.Service,
	uuidGen generator,
	clock clock,
) *Service {
	return &Service{
		stripeClient:    stripeClient,
		accountSvc:      accountSvc,
		subscriptionSvc: subscriptionSvc,
		invoiceSvc:      invoiceSvc,
		uuidGen:         uuidGen,
		clock:           clock,
	}
}

// CreateAccountWithStripe creates an account and a corresponding Stripe customer.
func (s *Service) CreateAccountWithStripe(
	ctx context.Context,
	accountName, email string,
) (account.Account, error) {
	cust, err := s.stripeClient.CreateCustomer(ctx, accountName, email)
	if err != nil {
		return account.Account{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"stripe create customer: %s",
			err,
		)
	}

	na := account.NewAccount{Name: accountName}

	created, err := s.accountSvc.Create(ctx, na)
	if err != nil {
		return account.Account{}, fmt.Errorf("create account: %w", err)
	}

	// Update with Stripe customer ID
	ua := account.UpdateAccount{
		StripeCustomerID: &sql.NullString{String: cust.ID, Valid: true},
	}

	updated, err := s.accountSvc.Update(ctx, created, ua)
	if err != nil {
		return account.Account{}, fmt.Errorf("update account stripe id: %w", err)
	}

	return updated, nil
}

// CreateCheckoutSession creates a Stripe Checkout session for an account.
func (s *Service) CreateCheckoutSession(
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

	acc, err := s.accountSvc.QueryByID(ctx, accountUUID)
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

	sess, err := s.stripeClient.CreateCheckoutSession(
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
func (s *Service) CreatePortalSession(
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

	acc, err := s.accountSvc.QueryByID(ctx, accountUUID)
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

	sess, err := s.stripeClient.CreateCustomerPortalSession(
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
func (s *Service) GetSubscriptions(
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

	subs, err := s.subscriptionSvc.QueryByAccountID(ctx, accountUUID)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}

	return toSubscriptionResponses(subs), nil
}

// GetInvoices returns invoices for an account.
func (s *Service) GetInvoices(ctx context.Context, accountID string) ([]InvoiceResponse, error) {
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"invalid account id: %s",
			err,
		)
	}

	invs, err := s.invoiceSvc.QueryByAccountID(ctx, accountUUID)
	if err != nil {
		return nil, fmt.Errorf("query invoices: %w", err)
	}

	return toInvoiceResponses(invs), nil
}

// HandleWebhookSubscription handles subscription webhook events.
func (s *Service) HandleWebhookSubscription(
	ctx context.Context,
	stripeSub *stripe.Subscription,
) error {
	existing, err := s.subscriptionSvc.QueryByStripeID(ctx, stripeSub.ID)
	if err != nil {
		// New subscription - create it
		return s.createSubscriptionFromStripe(ctx, stripeSub)
	}

	// Update existing
	existing = existing.WithStatus(string(stripeSub.Status)).
		WithCancelAtPeriodEnd(stripeSub.CancelAtPeriodEnd).
		WithUpdatedAt(s.clock.Now())

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

	updateErr := s.subscriptionSvc.Update(ctx, existing)
	if updateErr != nil {
		return fmt.Errorf("update subscription: %w", updateErr)
	}

	return nil
}

// HandleWebhookInvoice handles invoice webhook events.
func (s *Service) HandleWebhookInvoice(ctx context.Context, stripeInv *stripe.Invoice) error {
	existing, err := s.invoiceSvc.QueryByStripeID(ctx, stripeInv.ID)
	if err != nil {
		// New invoice - create it
		return s.createInvoiceFromStripe(ctx, stripeInv)
	}

	existing = existing.
		WithStatus(string(stripeInv.Status)).
		WithAmountPaid(stripeInv.AmountPaid).
		WithUpdatedAt(s.clock.Now())

	updateErr := s.invoiceSvc.Update(ctx, existing)
	if updateErr != nil {
		return fmt.Errorf("update invoice: %w", updateErr)
	}

	return nil
}

func (s *Service) createSubscriptionFromStripe(
	ctx context.Context,
	stripeSub *stripe.Subscription,
) error {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return fmt.Errorf("uuid error: %w", err)
	}

	// Look up account by stripe customer ID
	accountID, err := s.findAccountIDByStripeCustomer(ctx, stripeSub.Customer.ID)
	if err != nil {
		return fmt.Errorf("find account: %w", err)
	}

	priceID := ""
	if len(stripeSub.Items.Data) > 0 {
		priceID = stripeSub.Items.Data[0].Price.ID
	}

	now := s.clock.Now()

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

	createErr := s.subscriptionSvc.Create(ctx, sub)
	if createErr != nil {
		return fmt.Errorf("create subscription: %w", createErr)
	}

	return nil
}

func (s *Service) createInvoiceFromStripe(ctx context.Context, stripeInv *stripe.Invoice) error {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return fmt.Errorf("uuid error: %w", err)
	}

	accountID, err := s.findAccountIDByStripeCustomer(ctx, stripeInv.Customer.ID)
	if err != nil {
		return fmt.Errorf("find account: %w", err)
	}

	now := s.clock.Now()

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

	createErr := s.invoiceSvc.Create(ctx, inv)
	if createErr != nil {
		return fmt.Errorf("create invoice: %w", createErr)
	}

	return nil
}

func (s *Service) findAccountIDByStripeCustomer(
	ctx context.Context,
	stripeCustomerID string,
) (uuid.UUID, error) {
	filter := account.QueryFilter{
		StripeCustomerID: &stripeCustomerID,
	}

	accs, err := s.accountSvc.Query(ctx, filter, account.GetDefaultOrderBy(), defaultCursor())
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
