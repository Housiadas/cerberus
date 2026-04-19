package handler

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/core/user/user_search"
	"github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/sdk/pntr"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/mapper"
)

func (h *Handler) SearchUsers(
	ctx context.Context,
	request openapi.SearchUsersRequestObject,
) (openapi.SearchUsersResponseObject, error) {
	filter, sort, limit, searchAfter, err := parseSearchParams(request.Params)
	if err != nil {
		return nil, err
	}

	result, err := h.svc.user.Search(ctx, filter, sort, limit, searchAfter)
	if err != nil {
		if errors.Is(err, user.ErrSearchUnavailable) {
			return nil, errs.New(
				errs.Unavailable,
				errs.CodeSearchUnavailable,
				err,
			)
		}

		return nil, errs.New(errs.Internal, errs.CodeInternal, err)
	}

	return openapi.SearchUsers200JSONResponse{
		Data: mapper.MapSlice(result.Users, toOpenAPIUser),
		Metadata: openapi.SearchMetadata{
			TotalHits:   result.TotalHits,
			Limit:       limit,
			HasMore:     result.HasMore,
			SearchAfter: &result.SearchAfter,
		},
	}, nil
}

func parseSearchParams(
	params openapi.SearchUsersParams,
) (user_search.SearchFilter, string, int, string, error) {
	var fieldErrors errs.FieldErrors

	var filter user_search.SearchFilter

	filter.Q = pntr.DerefStr(params.Q)
	filter.Email = pntr.DerefStr(params.Email)
	filter.Department = pntr.DerefStr(params.Department)

	if params.Enabled != nil {
		enabled, err := strconv.ParseBool(*params.Enabled)
		if err != nil {
			fieldErrors.Add("enabled", err)
		} else {
			filter.Enabled = &enabled
		}
	}

	if params.StartCreatedDate != nil {
		t, err := time.Parse(time.RFC3339, *params.StartCreatedDate)
		if err != nil {
			fieldErrors.Add("start_created_date", err)
		} else {
			filter.StartCreatedDate = &t
		}
	}

	if params.EndCreatedDate != nil {
		t, err := time.Parse(time.RFC3339, *params.EndCreatedDate)
		if err != nil {
			fieldErrors.Add("end_created_date", err)
		} else {
			filter.EndCreatedDate = &t
		}
	}

	limit := 10
	if params.Limit != nil {
		l, err := strconv.Atoi(*params.Limit)
		if err != nil {
			fieldErrors.Add("limit", err)
		} else if l < 1 || l > 100 {
			fieldErrors.Add("limit", errors.New("must be between 1 and 100"))
		} else {
			limit = l
		}
	}

	if len(fieldErrors) > 0 {
		return user_search.SearchFilter{}, "", 0, "", fieldErrors.ToError()
	}

	sort := pntr.DerefStr(params.Sort)
	searchAfter := pntr.DerefStr(params.SearchAfter)

	return filter, sort, limit, searchAfter, nil
}
