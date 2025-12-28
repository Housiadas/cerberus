// Package permission_service provides internal access to permission core.
package permission_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Service manages the set of APIs for permission access.
type Service struct {
	log    *logger.Logger
	storer permission.Storer
}

// New constructor
func New(log *logger.Logger, storer permission.Storer) *Service {
	return &Service{log: log, storer: storer}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Service{
		log:    s.log,
		storer: storer,
	}

	return &bus, nil
}

// Create adds a new permission to the system.
func (s *Service) Create(ctx context.Context, np permission.NewPermission) (permission.Permission, error) {
	now := time.Now()
	p := permission.Permission{
		ID:        uuid.UUID{},
		Name:      np.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.storer.Create(ctx, p); err != nil {
		return permission.Permission{}, fmt.Errorf("permission create: %w", err)
	}

	return p, nil
}

// Update modifies information about a permission.
func (s *Service) Update(
	ctx context.Context,
	p permission.Permission,
	up permission.UpdatePermission,
) (permission.Permission, error) {
	if up.Name != nil {
		p.Name = *up.Name
	}

	p.UpdatedAt = time.Now()

	if err := s.storer.Update(ctx, p); err != nil {
		return permission.Permission{}, fmt.Errorf("permission update: %w", err)
	}

	return p, nil
}

// Delete removes the specified permission.
func (s *Service) Delete(ctx context.Context, p permission.Permission) error {
	if err := s.storer.Delete(ctx, p); err != nil {
		return fmt.Errorf("permission delete: %w", err)
	}

	return nil
}

// Count returns the total number of permissions.
func (s *Service) Count(ctx context.Context, filter permission.QueryFilter) (int, error) {
	return s.storer.Count(ctx, filter)
}

// QueryByID finds the permission by the specified ID.
func (s *Service) QueryByID(ctx context.Context, id uuid.UUID) (permission.Permission, error) {
	p, err := s.storer.QueryByID(ctx, id)
	if err != nil {
		return permission.Permission{}, fmt.Errorf("query: permissionID[%s]: %w", id, err)
	}

	return p, nil
}

// Query retrieves a list of existing permissions.
func (s *Service) Query(
	ctx context.Context,
	filter permission.QueryFilter,
	orderBy order.By,
	pg page.Page,
) ([]permission.Permission, error) {
	ps, err := s.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("permission query: %w", err)
	}

	return ps, nil
}
