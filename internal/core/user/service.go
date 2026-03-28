// Package user is the service of the user domain
package user

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// hasher defines the interface for password hashing operations.
type hasher interface {
	Hash(password string) ([]byte, error)
	Compare(hashedPassword []byte, password string) error
}

// dispatcher defines the interface for domain event dispatching.
type dispatcher interface {
	Dispatch(ctx context.Context, ev event.DomainEvent) error
}

// Service manages user domain operations including persistence,
// transaction management, and event dispatching.
type Service struct {
	log        logger.Logger
	storer     Storer
	uuidGen    generator
	clock      clock
	hasher     hasher
	db         *sqlx.DB
	dispatcher dispatcher
}

// NewService constructs the service.
func NewService(
	log logger.Logger,
	storer Storer,
	uuidGen generator,
	clock clock,
	hasher hasher,
	db *sqlx.DB,
	dispatcher dispatcher,
) *Service {
	return &Service{
		log:        log,
		storer:     storer,
		uuidGen:    uuidGen,
		clock:      clock,
		hasher:     hasher,
		db:         db,
		dispatcher: dispatcher,
	}
}

// Authenticate finds a user by their email and verifies their password. On
// success, it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (c *Service) Authenticate(
	ctx context.Context,
	email mail.Address,
	password string,
) (User, error) {
	usr, err := c.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	err = c.hasher.Compare(usr.PasswordHash(), password)
	if err != nil {
		return User{}, fmt.Errorf(
			"compare_hash_and_password: %w",
			ErrAuthenticationFailure,
		)
	}

	return usr, nil
}

// Create adds a new User to the system within a transaction
// and dispatches a UserCreated domain event.
func (c *Service) Create(ctx context.Context, nu NewUser) (User, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return User{}, fmt.Errorf("uuid v7 error: %w", err)
	}

	hash, err := c.hasher.Hash(nu.Password.String())
	if err != nil {
		return User{}, fmt.Errorf("generate_from_password: %w", err)
	}

	now := c.clock.Now()
	usr := New(id, nu.Name, nu.Email, hash, nu.Department, true, nil, now, now, nil)

	txErr := pgsql.RunInTx(ctx, c.log, c.db, func(txCtx context.Context) error {
		if err = c.storer.Create(txCtx, usr); err != nil {
			return fmt.Errorf("create: %w", err)
		}

		return c.dispatcher.Dispatch(txCtx, newUserEvent(usr, event.UserCreated, audit.ActionCreate))
	})
	if txErr != nil {
		return User{}, fmt.Errorf("create user: %w", txErr)
	}

	return usr, nil
}

// Query retrieves a list of existing users.
func (c *Service) Query(
	ctx context.Context,
	filter QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]User, error) {
	users, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

// QueryByID finds the user by the specified ID.
func (c *Service) QueryByID(ctx context.Context, userID uuid.UUID) (User, error) {
	usr, err := c.storer.QueryByID(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("query: userID[%s]: %w", userID, err)
	}

	return usr, nil
}

// QueryByEmail finds the user by a specified user email.
func (c *Service) QueryByEmail(ctx context.Context, email mail.Address) (User, error) {
	usr, err := c.storer.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	return usr, nil
}

// Update modifies information about a user.User within a transaction
// and dispatches a UserUpdated domain event.
func (c *Service) Update(
	ctx context.Context,
	usr User,
	uu UpdateUser,
) (User, error) {
	if uu.Name != nil {
		usr = usr.WithName(*uu.Name)
	}

	if uu.Email != nil {
		usr = usr.WithEmail(*uu.Email)
	}

	if uu.Password != nil {
		pw, err := c.hasher.Hash(uu.Password.String())
		if err != nil {
			return User{}, fmt.Errorf("generate_from_password: %w", err)
		}

		usr = usr.WithPasswordHash(pw)
	}

	if uu.Department != nil {
		usr = usr.WithDepartment(*uu.Department)
	}

	if uu.Enabled != nil {
		usr = usr.WithEnabled(*uu.Enabled)
	}

	usr = usr.WithUpdatedAt(c.clock.Now())

	txErr := pgsql.RunInTx(ctx, c.log, c.db, func(txCtx context.Context) error {
		if err := c.storer.Update(txCtx, usr); err != nil {
			return fmt.Errorf("update: %w", err)
		}

		return c.dispatcher.Dispatch(txCtx, newUserEvent(usr, event.UserUpdated, audit.ActionUpdate))
	})
	if txErr != nil {
		return User{}, fmt.Errorf("update user: %w", txErr)
	}

	return usr, nil
}

// Delete removes the specified user within a transaction
// and dispatches a UserDeleted domain event.
func (c *Service) Delete(ctx context.Context, usr User) error {
	txErr := pgsql.RunInTx(ctx, c.log, c.db, func(txCtx context.Context) error {
		if err := c.storer.Delete(txCtx, usr); err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		return c.dispatcher.Dispatch(txCtx, newUserEvent(usr, event.UserDeleted, audit.ActionDelete))
	})
	if txErr != nil {
		return fmt.Errorf("delete user: %w", txErr)
	}

	return nil
}

// VerifyPassword checks whether the given plaintext password matches the
// user's stored password hash. Returns user.ErrAuthenticationFailure on mismatch.
func (c *Service) VerifyPassword(usr User, plainPassword string) error {
	err := c.hasher.Compare(usr.PasswordHash(), plainPassword)
	if err != nil {
		return ErrAuthenticationFailure
	}

	return nil
}

// newUserEvent creates a DomainEvent for user operations.
func newUserEvent(usr User, eventType string, action string) event.DomainEvent {
	return event.DomainEvent{
		EventType:   eventType,
		AggregateID: usr.ID(),
		Topic:       event.UserTopic,
		Payload:     usr,
		ObjEntity:   entity.New(entity.UserEntity),
		ObjName:     usr.Name(),
		Action:      action,
		Message:     "user " + action,
	}
}
