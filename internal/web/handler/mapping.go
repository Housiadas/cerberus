package handler

import (
	"bytes"
	"encoding/json"

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
		Id:        r.ID().String(),
		Name:      r.Name().String(),
		CreatedAt: clock.Format(new(r.CreatedAt())),
		UpdatedAt: clock.Format(new(r.UpdatedAt())),
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
		Id:        p.ID().String(),
		Name:      p.Name().String(),
		CreatedAt: clock.Format(new(p.CreatedAt())),
		UpdatedAt: clock.Format(new(p.UpdatedAt())),
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
		Id:               acc.ID().String(),
		Name:             acc.Name(),
		StripeCustomerId: acc.StripeCustomerID().String,
		CreatedAt:        clock.Format(new(acc.CreatedAt())),
		UpdatedAt:        clock.Format(new(acc.UpdatedAt())),
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
		Id:        a.ID().String(),
		ObjId:     a.ObjID().String(),
		ObjEntity: a.ObjEntity().String(),
		ObjName:   a.ObjName().String(),
		ActorId:   a.ActorID().String(),
		Action:    a.Action(),
		Data:      compactJSON(a.Data()),
		Message:   a.Message(),
		Timestamp: clock.Format(new(a.Timestamp())),
	}
}

func compactJSON(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var buf bytes.Buffer
	err := json.Compact(&buf, data)
	if err != nil {
		return string(data)
	}

	return buf.String()
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
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresIn:    t.ExpiresIn,
	}
}

func toOpenAPIMetadata(m cursor.Metadata) openapi.Metadata {
	return openapi.Metadata{
		NextCursor: m.NextCursor,
		PrevCursor: m.PrevCursor,
		HasMore:    m.HasMore,
		Limit:      m.Limit,
	}
}

func toOpenAPICheckoutResponse(r billing.CheckoutResponse) openapi.CheckoutResponse {
	return openapi.CheckoutResponse{
		Url: r.URL,
	}
}

func toOpenAPIPortalResponse(r billing.PortalResponse) openapi.PortalResponse {
	return openapi.PortalResponse{
		Url: r.URL,
	}
}

func toOpenAPISubscriptionResponse(r billing.SubscriptionResponse) openapi.SubscriptionResponse {
	return openapi.SubscriptionResponse{
		Id:                   r.ID,
		StripeSubscriptionId: r.StripeSubscriptionID,
		StripePriceId:        r.StripePriceID,
		Status:               r.Status,
		CurrentPeriodStart:   r.CurrentPeriodStart,
		CurrentPeriodEnd:     r.CurrentPeriodEnd,
		CancelAtPeriodEnd:    r.CancelAtPeriodEnd,
		CreatedAt:            r.CreatedAt,
	}
}

func toOpenAPISubscriptionResponses(
	subs []billing.SubscriptionResponse,
) []openapi.SubscriptionResponse {
	out := make([]openapi.SubscriptionResponse, len(subs))
	for i, s := range subs {
		out[i] = toOpenAPISubscriptionResponse(s)
	}

	return out
}

func toOpenAPIInvoiceResponse(r billing.InvoiceResponse) openapi.InvoiceResponse {
	return openapi.InvoiceResponse{
		Id:              r.ID,
		StripeInvoiceId: r.StripeInvoiceID,
		Status:          r.Status,
		AmountDue:       r.AmountDue,
		AmountPaid:      r.AmountPaid,
		Currency:        r.Currency,
		InvoiceUrl:      r.InvoiceURL,
		CreatedAt:       r.CreatedAt,
	}
}

func toOpenAPIInvoiceResponses(invs []billing.InvoiceResponse) []openapi.InvoiceResponse {
	out := make([]openapi.InvoiceResponse, len(invs))
	for i, inv := range invs {
		out[i] = toOpenAPIInvoiceResponse(inv)
	}

	return out
}
