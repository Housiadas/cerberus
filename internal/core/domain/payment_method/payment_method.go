package payment_method

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("payment method not found")

type PaymentMethod struct {
	id        uuid.UUID
	accountID uuid.UUID
	typ       PaymentMethodType
	details   string
	isDefault bool
	createdAt time.Time
	updatedAt time.Time
}

func New(
	id uuid.UUID,
	accountID uuid.UUID,
	typ PaymentMethodType,
	details string,
	isDefault bool,
	createdAt time.Time,
	updatedAt time.Time,
) PaymentMethod {
	return PaymentMethod{
		id: id, accountID: accountID, typ: typ, details: details,
		isDefault: isDefault, createdAt: createdAt, updatedAt: updatedAt,
	}
}

func (p PaymentMethod) ID() uuid.UUID           { return p.id }
func (p PaymentMethod) AccountID() uuid.UUID    { return p.accountID }
func (p PaymentMethod) Type() PaymentMethodType { return p.typ }
func (p PaymentMethod) Details() string         { return p.details }
func (p PaymentMethod) IsDefault() bool         { return p.isDefault }
func (p PaymentMethod) CreatedAt() time.Time    { return p.createdAt }
func (p PaymentMethod) UpdatedAt() time.Time    { return p.updatedAt }

type NewPaymentMethod struct {
	AccountID uuid.UUID
	Type      PaymentMethodType
	Details   string
	IsDefault bool
}

type QueryFilter struct {
	ID        *uuid.UUID
	AccountID *uuid.UUID
	Type      *PaymentMethodType
}
