// Package user_service it is the service of the user domain
package user

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
}

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// Hasher defines the interface for password hashing operations.
type hasher interface {
	Hash(password string) ([]byte, error)
	Compare(hashedPassword []byte, password string) error
}

type Service struct {
	log     logger
	storer  Storer
	uuidGen generator
	clock   clock
	hasher  hasher
}

// NewService constructs the service.
func NewService(
	log logger,
	storer Storer,
	uuidGen generator,
	clock clock,
	hasher hasher,
) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
		clock:   clock,
		hasher:  hasher,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (c *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := c.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("user transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
		clock:   c.clock,
		hasher:  c.hasher,
	}

	return &bus, nil
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

// Create adds a new User to the system.
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

	err = c.storer.Create(ctx, usr)
	if err != nil {
		return User{}, fmt.Errorf("create: %w", err)
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

// Update modifies information about a user.User.
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

	err := c.storer.Update(ctx, usr)
	if err != nil {
		return User{}, fmt.Errorf("update: %w", err)
	}

	return usr, nil
}

// Delete removes the specified user.
func (c *Service) Delete(ctx context.Context, usr User) error {
	err := c.storer.Delete(ctx, usr)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
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
