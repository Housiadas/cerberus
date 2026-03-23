package billing_usecase

// CheckoutRequest defines the data needed to create a Stripe Checkout session.
type CheckoutRequest struct {
	AccountID  string `json:"accountId"  validate:"required"`
	PriceID    string `json:"priceId"    validate:"required"`
	SuccessURL string `json:"successUrl" validate:"required"`
	CancelURL  string `json:"cancelUrl"  validate:"required"`
}

// CheckoutResponse contains the Stripe Checkout session URL.
type CheckoutResponse struct {
	URL string `json:"url"`
}

// PortalRequest defines the data needed to create a Stripe Customer Portal session.
type PortalRequest struct {
	AccountID string `json:"accountId" validate:"required"`
	ReturnURL string `json:"returnUrl" validate:"required"`
}

// PortalResponse contains the Stripe Customer Portal session URL.
type PortalResponse struct {
	URL string `json:"url"`
}

// SubscriptionResponse represents a subscription for the API.
type SubscriptionResponse struct {
	ID                   string  `json:"id"`
	StripeSubscriptionID string  `json:"stripeSubscriptionId"`
	StripePriceID        string  `json:"stripePriceId"`
	Status               string  `json:"status"`
	CurrentPeriodStart   *string `json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd     *string `json:"currentPeriodEnd,omitempty"`
	CancelAtPeriodEnd    bool    `json:"cancelAtPeriodEnd"`
	CreatedAt            string  `json:"createdAt"`
}

// InvoiceResponse represents an invoice for the API.
type InvoiceResponse struct {
	ID              string `json:"id"`
	StripeInvoiceID string `json:"stripeInvoiceId"`
	Status          string `json:"status"`
	AmountDue       int64  `json:"amountDue"`
	AmountPaid      int64  `json:"amountPaid"`
	Currency        string `json:"currency"`
	InvoiceURL      string `json:"invoiceUrl,omitempty"`
	CreatedAt       string `json:"createdAt"`
}
