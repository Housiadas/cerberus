// Package transaction_usecase maintains the cli layer http for the tran core.
package transaction_usecase

import (
	"context"
	"errors"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/product_core"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// App manages the set of cli layer http functions for the tran core.
type App struct {
	userBus    *user_service.Service
	productBus *product_core.Core
}

// NewApp constructs a tran cli API for use.
func NewApp(userBus *user_service.Service, productBus *product_core.Core) *App {
	return &App{
		userBus:    userBus,
		productBus: productBus,
	}
}

// newWithTx constructs a new Handlers value with the core apis
// using a store transaction that was created via middleware.
func (a *App) newWithTx(ctx context.Context) (*App, error) {
	tx, err := pgsql.GetTran(ctx)
	if err != nil {
		return nil, err
	}

	userBus, err := a.userBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	productBus, err := a.productBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	app := App{
		userBus:    userBus,
		productBus: productBus,
	}

	return &app, nil
}

// Create adds a new user and product at the same time under a single transaction.
func (a *App) Create(ctx context.Context, nt NewTran) (Product, error) {
	a, err := a.newWithTx(ctx)
	if err != nil {
		return Product{}, errs.New(errs.Internal, err)
	}

	np, err := toBusNewProduct(nt.Product)
	if err != nil {
		return Product{}, errs.New(errs.InvalidArgument, err)
	}

	nu, err := toBusNewUser(nt.User)
	if err != nil {
		return Product{}, errs.New(errs.InvalidArgument, err)
	}

	usr, err := a.userBus.Create(ctx, nu)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return Product{}, errs.New(errs.Aborted, user.ErrUniqueEmail)
		}
		return Product{}, errs.Newf(errs.Internal, "create: usr[%+v]: %s", usr, err)
	}

	np.UserID = usr.ID

	prd, err := a.productBus.Create(ctx, np)
	if err != nil {
		return Product{}, errs.Newf(errs.Internal, "create: prd[%+v]: %s", prd, err)
	}

	return toAppProduct(prd), nil
}
