package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/sdk/pntr"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

func (h *Handler) CreateAccount(
	ctx context.Context,
	request openapi.CreateAccountRequestObject,
) (openapi.CreateAccountResponseObject, error) {
	acc, err := h.svc.account.Create(ctx, account.NewAccount{
		Name: request.Body.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	return openapi.CreateAccount200JSONResponse(toOpenAPIAccount(acc)), nil
}

//nolint:dupl // list handler pattern shared across domains
func (h *Handler) ListAccounts(
	ctx context.Context,
	request openapi.ListAccountsRequestObject,
) (openapi.ListAccountsResponseObject, error) {
	cur, err := cursor.Parse(
		pntr.DerefStr(request.Params.Cursor),
		pntr.DerefStr(request.Params.Limit),
	)
	if err != nil {
		return nil, errs.NewFieldErrors("cursor", err)
	}

	filter, err := parseAccountFilter(
		pntr.DerefStr(request.Params.AccountId),
		pntr.DerefStr(request.Params.Name),
	)
	if err != nil {
		return nil, err
	}

	orderBy, err := order.Parse(
		accountOrderByFields(),
		pntr.DerefStr(request.Params.OrderBy),
		accountDefaultOrderBy(),
	)
	if err != nil {
		return nil, errs.NewFieldErrors("order", err)
	}

	accs, err := h.svc.account.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "query: %s", err)
	}

	result := cursor.NewResult(accs, cur.Limit(), cur, orderBy,
		func(a account.Account) string { return a.ID().String() },
		accountFieldExtractor(orderBy),
	)

	return openapi.ListAccounts200JSONResponse{
		Data:     new(toOpenAPIAccounts(result.Data)),
		Metadata: new(toOpenAPIMetadata(result.Metadata)),
	}, nil
}

func (h *Handler) GetAccount(
	ctx context.Context,
	request openapi.GetAccountRequestObject,
) (openapi.GetAccountResponseObject, error) {
	accountUUID, err := uuid.Parse(request.AccountId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	acc, err := h.svc.account.QueryByID(ctx, accountUUID)
	if err != nil {
		if errors.Is(err, account.ErrNotFound) {
			return nil, errs.New(errs.NotFound, errs.CodeAccountNotFound, err)
		}

		return nil, fmt.Errorf("get account: %w", err)
	}

	return openapi.GetAccount200JSONResponse(toOpenAPIAccount(acc)), nil
}

func (h *Handler) UpdateAccount(
	ctx context.Context,
	request openapi.UpdateAccountRequestObject,
) (openapi.UpdateAccountResponseObject, error) {
	accountUUID, err := uuid.Parse(request.AccountId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentAcc, err := h.svc.account.QueryByID(ctx, accountUUID)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "query by id: %s", err)
	}

	ua := account.UpdateAccount{
		Name: request.Body.Name,
	}

	updAcc, err := h.svc.account.Update(ctx, currentAcc, ua)
	if err != nil {
		return nil, fmt.Errorf("update account: %w", err)
	}

	return openapi.UpdateAccount200JSONResponse(toOpenAPIAccount(updAcc)), nil
}

func (h *Handler) DeleteAccount(
	ctx context.Context,
	request openapi.DeleteAccountRequestObject,
) (openapi.DeleteAccountResponseObject, error) {
	accountUUID, err := uuid.Parse(request.AccountId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentAcc, err := h.svc.account.QueryByID(ctx, accountUUID)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "query by id: %s", err)
	}

	err = h.svc.account.Delete(ctx, currentAcc)
	if err != nil {
		return nil, fmt.Errorf("delete account: %w", err)
	}

	return openapi.DeleteAccount204Response{}, nil
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

func parseAccountFilter(id, nameStr string) (account.QueryFilter, error) {
	var (
		fieldErrors errs.FieldErrors
		filter      account.QueryFilter
	)

	if id != "" {
		uid, err := uuid.Parse(id)
		if err != nil {
			fieldErrors.Add("account_id", err)
		} else {
			filter.ID = &uid
		}
	}

	if nameStr != "" {
		filter.Name = &nameStr
	}

	if fieldErrors != nil {
		return account.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}

func accountDefaultOrderBy() order.By {
	return order.NewBy("account_id", order.ASC)
}

func accountOrderByFields() map[string]string {
	return map[string]string{
		"account_id": account.OrderByID,
		"name":       account.OrderByName,
	}
}

func accountFieldExtractor(orderBy order.By) func(account.Account) any {
	return func(a account.Account) any {
		switch orderBy.Field {
		case account.OrderByName:
			return a.Name()
		default:
			return a.ID().String()
		}
	}
}
