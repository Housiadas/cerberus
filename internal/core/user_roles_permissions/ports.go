package user_roles_permissions

import (
	"context"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
)

// storer interface declares the behavior this package needs to retrieve data from the view.
type storer interface {
	GetPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]db.GetPermissionsByUserIDRow, error)
}
