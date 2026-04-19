package audit

import (
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateAuditParams(
	id uuid.UUID,
	na NewAudit,
	jsonData []byte,
	now time.Time,
) db.CreateAuditParams {
	return db.CreateAuditParams{
		ID:        id,
		ObjID:     na.ObjID,
		ObjEntity: na.ObjEntity.String(),
		ObjName:   na.ObjName.String(),
		ActorID:   na.ActorID,
		Action:    na.Action,
		Data:      jsonData,
		Message:   pgtype.Text{String: na.Message, Valid: na.Message != ""},
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
	}
}

// toDomainAudit converts a db.Audit to the domain Audit type.
func toDomainAudit(a db.Audit) Audit {
	return New(
		a.ID,
		a.ObjID,
		entity.MustParse(a.ObjEntity),
		name.MustParse(a.ObjName),
		a.ActorID,
		a.Action,
		a.Data,
		a.Message.String,
		a.CreatedAt.Time,
	)
}

// toDBQueryFilter converts domain QueryFilter to db.AuditQueryFilter.
func toDBQueryFilter(f QueryFilter) db.AuditQueryFilter {
	dbf := db.AuditQueryFilter{
		ObjID:   f.ObjID,
		ActorID: f.ActorID,
		Action:  f.Action,
		Since:   f.Since,
		Until:   f.Until,
	}

	if f.ObjEntity != nil {
		dbf.ObjEntity = new(f.ObjEntity.String())
	}

	if f.ObjName != nil {
		dbf.ObjName = new(f.ObjName.String())
	}

	return dbf
}
