package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/pntr"
	"github.com/Housiadas/cerberus/internal/usecase/account_usecase"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

func (h *Handler) CreateAccount(
	ctx context.Context,
	request openapi.CreateAccountRequestObject,
) (openapi.CreateAccountResponseObject, error) {
	acc, err := h.usecase.account.Create(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	return openapi.CreateAccount200JSONResponse(acc), nil
}

func (h *Handler) ListAccounts(
	ctx context.Context,
	request openapi.ListAccountsRequestObject,
) (openapi.ListAccountsResponseObject, error) {
	qp := account_usecase.AppQueryParams{
		Cursor:  pntr.DerefStr(request.Params.Cursor),
		Limit:   pntr.DerefStr(request.Params.Limit),
		OrderBy: pntr.DerefStr(request.Params.OrderBy),
		ID:      pntr.DerefStr(request.Params.AccountId),
		Name:    pntr.DerefStr(request.Params.Name),
	}

	result, err := h.usecase.account.Query(ctx, qp)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	return openapi.ListAccounts200JSONResponse{
		Data:     new(result.Data),
		Metadata: new(result.Metadata),
	}, nil
}

func (h *Handler) GetAccount(
	ctx context.Context,
	request openapi.GetAccountRequestObject,
) (openapi.GetAccountResponseObject, error) {
	acc, err := h.usecase.account.QueryByID(ctx, request.AccountId)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	return openapi.GetAccount200JSONResponse(acc), nil
}

func (h *Handler) UpdateAccount(
	ctx context.Context,
	request openapi.UpdateAccountRequestObject,
) (openapi.UpdateAccountResponseObject, error) {
	acc, err := h.usecase.account.Update(ctx, *request.Body, request.AccountId)
	if err != nil {
		return nil, fmt.Errorf("update account: %w", err)
	}

	return openapi.UpdateAccount200JSONResponse(acc), nil
}

func (h *Handler) DeleteAccount(
	ctx context.Context,
	request openapi.DeleteAccountRequestObject,
) (openapi.DeleteAccountResponseObject, error) {
	err := h.usecase.account.Delete(ctx, request.AccountId)
	if err != nil {
		return nil, fmt.Errorf("delete account: %w", err)
	}

	return openapi.DeleteAccount204Response{}, nil
}
