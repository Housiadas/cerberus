package handler

import (
	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/auth"
	"github.com/Housiadas/cerberus/internal/core/billing"
	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/cursor"
)

// ---------------------------------------------------------------------------
// Domain → OpenAPI mappers
// ---------------------------------------------------------------------------

func toOpenAPIUser(u user.User) openapi.User {
	return openapi.User{
		Id:         u.ID().String(),
		Name:       u.Name().String(),
		Email:      u.Email().Address,
		Department: u.Department().String(),
		Enabled:    u.Enabled(),
		CreatedAt:  clock.Format(new(u.CreatedAt())),
		UpdatedAt:  clock.Format(new(u.UpdatedAt())),
	}
}

func toOpenAPIRole(r role.Role) openapi.Role {
	return openapi.Role{
		Id:        new(r.ID().String()),
		Name:      new(r.Name().String()),
		CreatedAt: new(clock.Format(new(r.CreatedAt()))),
		UpdatedAt: new(clock.Format(new(r.UpdatedAt()))),
	}
}

func toOpenAPIRoles(roles []role.Role) []openapi.Role {
	out := make([]openapi.Role, len(roles))
	for i, r := range roles {
		out[i] = toOpenAPIRole(r)
	}

	return out
}

func toOpenAPIPermission(p permission.Permission) openapi.Permission {
	return openapi.Permission{
		Id:        new(p.ID().String()),
		Name:      new(p.Name().String()),
		CreatedAt: new(clock.Format(new(p.CreatedAt()))),
		UpdatedAt: new(clock.Format(new(p.UpdatedAt()))),
	}
}

func toOpenAPIPermissions(perms []permission.Permission) []openapi.Permission {
	out := make([]openapi.Permission, len(perms))
	for i, p := range perms {
		out[i] = toOpenAPIPermission(p)
	}

	return out
}

func toOpenAPIAccount(acc account.Account) openapi.Account {
	return openapi.Account{
		Id:               new(acc.ID().String()),
		Name:             new(acc.Name()),
		StripeCustomerId: new(acc.StripeCustomerID().String),
		CreatedAt:        new(clock.Format(new(acc.CreatedAt()))),
		UpdatedAt:        new(clock.Format(new(acc.UpdatedAt()))),
	}
}

func toOpenAPIAccounts(accs []account.Account) []openapi.Account {
	out := make([]openapi.Account, len(accs))
	for i, acc := range accs {
		out[i] = toOpenAPIAccount(acc)
	}

	return out
}

func toOpenAPIAudit(a audit.Audit) openapi.Audit {
	return openapi.Audit{
		Id:        new(a.ID().String()),
		ObjId:     new(a.ObjID().String()),
		ObjEntity: new(a.ObjEntity().String()),
		ObjName:   new(a.ObjName().String()),
		ActorId:   new(a.ActorID().String()),
		Action:    new(a.Action()),
		Data:      new(string(a.Data())),
		Message:   new(a.Message()),
		Timestamp: new(clock.Format(new(a.Timestamp()))),
	}
}

func toOpenAPIAudits(audits []audit.Audit) []openapi.Audit {
	out := make([]openapi.Audit, len(audits))
	for i, a := range audits {
		out[i] = toOpenAPIAudit(a)
	}

	return out
}

func toOpenAPIToken(t auth.Token) openapi.Token {
	return openapi.Token{
		AccessToken:  new(t.AccessToken),
		RefreshToken: new(t.RefreshToken),
		ExpiresIn:    new(t.ExpiresIn),
	}
}

func toOpenAPIMetadata(m cursor.Metadata) openapi.Metadata {
	return openapi.Metadata{
		NextCursor: new(m.NextCursor),
		PrevCursor: new(m.PrevCursor),
		HasMore:    new(m.HasMore),
		Limit:      new(m.Limit),
	}
}

func toOpenAPICheckoutResponse(r billing.CheckoutResponse) openapi.CheckoutResponse {
	return openapi.CheckoutResponse{
		Url: new(r.URL),
	}
}

func toOpenAPIPortalResponse(r billing.PortalResponse) openapi.PortalResponse {
	return openapi.PortalResponse{
		Url: new(r.URL),
	}
}

func toOpenAPISubscriptionResponse(r billing.SubscriptionResponse) openapi.SubscriptionResponse {
	return openapi.SubscriptionResponse{
		Id:                   new(r.ID),
		StripeSubscriptionId: new(r.StripeSubscriptionID),
		StripePriceId:        new(r.StripePriceID),
		Status:               new(r.Status),
		CurrentPeriodStart:   r.CurrentPeriodStart,
		CurrentPeriodEnd:     r.CurrentPeriodEnd,
		CancelAtPeriodEnd:    new(r.CancelAtPeriodEnd),
		CreatedAt:            new(r.CreatedAt),
	}
}

func toOpenAPISubscriptionResponses(subs []billing.SubscriptionResponse) []openapi.SubscriptionResponse {
	out := make([]openapi.SubscriptionResponse, len(subs))
	for i, s := range subs {
		out[i] = toOpenAPISubscriptionResponse(s)
	}

	return out
}

func toOpenAPIInvoiceResponse(r billing.InvoiceResponse) openapi.InvoiceResponse {
	return openapi.InvoiceResponse{
		Id:              new(r.ID),
		StripeInvoiceId: new(r.StripeInvoiceID),
		Status:          new(r.Status),
		AmountDue:       new(r.AmountDue),
		AmountPaid:      new(r.AmountPaid),
		Currency:        new(r.Currency),
		InvoiceUrl:      new(r.InvoiceURL),
		CreatedAt:       new(r.CreatedAt),
	}
}

func toOpenAPIInvoiceResponses(invs []billing.InvoiceResponse) []openapi.InvoiceResponse {
	out := make([]openapi.InvoiceResponse, len(invs))
	for i, inv := range invs {
		out[i] = toOpenAPIInvoiceResponse(inv)
	}

	return out
}
