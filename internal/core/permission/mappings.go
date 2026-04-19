package permission

import (
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreatePermissionParams(
	id uuid.UUID,
	np NewPermission,
	now time.Time,
) db.CreatePermissionParams {
	return db.CreatePermissionParams{
		ID:        id,
		Name:      np.Name.String(),
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: now, Valid: true},
	}
}

func toUpdatePermissionParams(p Permission) db.UpdatePermissionParams {
	return db.UpdatePermissionParams{
		ID:        p.ID(),
		Name:      pgtype.Text{String: p.Name().String(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: p.UpdatedAt(), Valid: true},
	}
}

func toDBQueryFilter(f QueryFilter) db.PermissionQueryFilter {
	dbf := db.PermissionQueryFilter{
		ID: f.ID,
	}

	if f.Name != nil {
		s := f.Name.String()
		dbf.Name = &s
	}

	return dbf
}

func toDomainPermission(p db.Permission) Permission {
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}

	return New(
		p.ID,
		name.MustParse(p.Name),
		p.CreatedAt.Time,
		p.UpdatedAt.Time,
		deletedAt,
	)
}
