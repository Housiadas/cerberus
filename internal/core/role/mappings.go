package role

import (
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateRoleParams(id uuid.UUID, nr NewRole, now time.Time) db.CreateRoleParams {
	return db.CreateRoleParams{
		ID:        id,
		Name:      nr.Name.String(),
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: now, Valid: true},
	}
}

func toUpdateRoleParams(rl Role) db.UpdateRoleParams {
	return db.UpdateRoleParams{
		ID:        rl.ID(),
		Name:      pgtype.Text{String: rl.Name().String(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: rl.UpdatedAt(), Valid: true},
	}
}

func toDBQueryFilter(f QueryFilter) db.RoleQueryFilter {
	dbf := db.RoleQueryFilter{
		ID: f.ID,
	}

	if f.Name != nil {
		s := f.Name.String()
		dbf.Name = &s
	}

	return dbf
}

func toDomainRole(r db.Role) Role {
	var deletedAt *time.Time
	if r.DeletedAt.Valid {
		deletedAt = &r.DeletedAt.Time
	}

	return New(
		r.ID,
		name.MustParse(r.Name),
		r.CreatedAt.Time,
		r.UpdatedAt.Time,
		deletedAt,
	)
}

func toDomainRoles(dbRoles []db.Role) []Role {
	roles := make([]Role, len(dbRoles))
	for i, r := range dbRoles {
		roles[i] = toDomainRole(r)
	}

	return roles
}

func toDomainRoleFromGetByID(r db.GetRoleByIDRow) Role {
	return New(
		r.ID,
		name.MustParse(r.Name),
		r.CreatedAt.Time,
		r.UpdatedAt.Time,
		nil,
	)
}
