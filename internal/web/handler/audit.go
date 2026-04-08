package handler

import (
	"context"
	"time"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/sdk/pntr"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

func (h *Handler) ListAudits(
	ctx context.Context,
	request openapi.ListAuditsRequestObject,
) (openapi.ListAuditsResponseObject, error) {
	cur, err := cursor.Parse(
		pntr.DerefStr(request.Params.Cursor),
		pntr.DerefStr(request.Params.Limit),
	)
	if err != nil {
		return nil, errs.NewFieldErrors("cursor", err)
	}

	filter, err := parseAuditFilter(
		pntr.DerefStr(request.Params.ObjId),
		pntr.DerefStr(request.Params.ObjDomain),
		pntr.DerefStr(request.Params.ObjName),
		pntr.DerefStr(request.Params.ActorId),
		pntr.DerefStr(request.Params.Action),
		pntr.DerefStr(request.Params.Since),
		pntr.DerefStr(request.Params.Until),
	)
	if err != nil {
		return nil, err
	}

	orderBy, err := order.Parse(
		auditOrderByFields(),
		pntr.DerefStr(request.Params.OrderBy),
		auditDefaultOrderBy(),
	)
	if err != nil {
		return nil, errs.NewFieldErrors("order", err)
	}

	audits, err := h.svc.audit.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "query: %s", err)
	}

	result := cursor.NewResult(audits, cur.Limit(), cur, orderBy,
		func(a audit.Audit) string { return a.ID().String() },
		auditFieldExtractor(orderBy),
	)

	return openapi.ListAudits200JSONResponse{
		Data:     new(toOpenAPIAudits(result.Data)),
		Metadata: new(toOpenAPIMetadata(result.Metadata)),
	}, nil
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

//nolint:cyclop
func parseAuditFilter(
	objID, objEntity, objName, actorID, action, since, until string,
) (audit.QueryFilter, error) {
	var (
		fieldErrors errs.FieldErrors
		filter      audit.QueryFilter
	)

	if objID != "" {
		id, err := uuid.Parse(objID)
		if err != nil {
			fieldErrors.Add("obj_id", err)
		} else {
			filter.ObjID = &id
		}
	}

	if objEntity != "" {
		domain, err := entity.Parse(objEntity)
		if err != nil {
			fieldErrors.Add("obj_entity", err)
		} else {
			filter.ObjEntity = &domain
		}
	}

	if objName != "" {
		n, err := name.Parse(objName)
		if err != nil {
			fieldErrors.Add("obj_name", err)
		} else {
			filter.ObjName = &n
		}
	}

	if actorID != "" {
		id, err := uuid.Parse(actorID)
		if err != nil {
			fieldErrors.Add("actor_id", err)
		} else {
			filter.ActorID = &id
		}
	}

	if action != "" {
		filter.Action = &action
	}

	if since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err != nil {
			fieldErrors.Add("since", err)
		} else {
			filter.Since = &t
		}
	}

	if until != "" {
		t, err := time.Parse(time.RFC3339, until)
		if err != nil {
			fieldErrors.Add("until", err)
		} else {
			filter.Until = &t
		}
	}

	if len(fieldErrors) > 0 {
		return audit.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}

func auditDefaultOrderBy() order.By {
	return order.NewBy("obj_id", order.ASC)
}

func auditOrderByFields() map[string]string {
	return map[string]string{
		"obj_id":     audit.OrderByObjID,
		"obj_entity": audit.OrderByObjEntity,
		"obj_name":   audit.OrderByObjName,
		"actor_id":   audit.OrderByActorID,
		"action":     audit.OrderByAction,
	}
}

func auditFieldExtractor(orderBy order.By) func(audit.Audit) any {
	return func(a audit.Audit) any {
		switch orderBy.Field {
		case audit.OrderByObjID:
			return a.ObjID().String()
		case audit.OrderByObjEntity:
			return a.ObjEntity().String()
		case audit.OrderByObjName:
			return a.ObjName().String()
		case audit.OrderByActorID:
			return a.ActorID().String()
		case audit.OrderByAction:
			return a.Action()
		default:
			return a.ID().String()
		}
	}
}
