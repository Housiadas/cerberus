package entity

const (
	UserEntity                 = "USER"
	RoleEntity                 = "ROLE"
	PermissionEntity           = "PERMISSION"
	AccountEntity              = "ACCOUNT"
	PlanEntity                 = "PLAN"
	SubscriptionEntity         = "SUBSCRIPTION"
	PaymentMethodEntity        = "PAYMENT_METHOD"
	InvoiceEntity              = "INVOICE"
	InvoiceItemEntity          = "INVOICE_ITEM"
	PaymentEntity              = "PAYMENT"
	UsageRecordEntity          = "USAGE_RECORD"
	CouponEntity               = "COUPON"
	SubscriptionDiscountEntity = "SUBSCRIPTION_DISCOUNT"
	TaxRateEntity              = "TAX_RATE"
	RefundEntity               = "REFUND"
	BillingAddressEntity       = "BILLING_ADDRESS"
)

// Entity represents a domain in the system.
type Entity struct {
	value string
}

func New(entity string) Entity {
	return Entity{entity}
}

// String returns the name of the role.
func (e Entity) String() string {
	return e.value
}

// Equal provides support for the go-cmp package and testing.
func (e Entity) Equal(d2 Entity) bool {
	return e.value == d2.value
}

// MarshalText provides support for logging and any marshal needs.
func (e Entity) MarshalText() ([]byte, error) {
	return []byte(e.value), nil
}

func getEntities() map[string]Entity {
	return map[string]Entity{
		UserEntity:                 New("USER"),
		RoleEntity:                 New("ROLE"),
		PermissionEntity:           New("PERMISSION"),
		AccountEntity:              New("ACCOUNT"),
		PlanEntity:                 New("PLAN"),
		SubscriptionEntity:         New("SUBSCRIPTION"),
		PaymentMethodEntity:        New("PAYMENT_METHOD"),
		InvoiceEntity:              New("INVOICE"),
		InvoiceItemEntity:          New("INVOICE_ITEM"),
		PaymentEntity:              New("PAYMENT"),
		UsageRecordEntity:          New("USAGE_RECORD"),
		CouponEntity:               New("COUPON"),
		SubscriptionDiscountEntity: New("SUBSCRIPTION_DISCOUNT"),
		TaxRateEntity:              New("TAX_RATE"),
		RefundEntity:               New("REFUND"),
		BillingAddressEntity:       New("BILLING_ADDRESS"),
	}
}
