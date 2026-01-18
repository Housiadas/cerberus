package user_service

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"
)

// Query retrieves a list of existing users.
func (c *Service) Query(
	ctx context.Context,
	filter user.QueryFilter,
	orderBy order.By,
	page web.Page,
) ([]user.User, error) {
	users, err := c.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

// QueryByID finds the user by the specified ID.
func (c *Service) QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	usr, err := c.storer.QueryByID(ctx, userID)
	if err != nil {
		return user.User{}, fmt.Errorf("query: userID[%s]: %w", userID, err)
	}

	return usr, nil
}

// QueryByEmail finds the user by a specified user email.
func (c *Service) QueryByEmail(ctx context.Context, email mail.Address) (user.User, error) {
	usr, err := c.storer.QueryByEmail(ctx, email)
	if err != nil {
		return user.User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	return usr, nil
}
