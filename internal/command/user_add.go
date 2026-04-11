package command

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/sdk/eventbus"
	namePck "github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/internal/types/password"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

// UserAdd adds new users into the database.
func (cmd *Command) UserAdd(name, email, pass string) error {
	if name == "" || email == "" || pass == "" {
		fmt.Println("help: useradd <name> <email> <password>")

		return ErrHelp
	}

	dbPool, err := db.Open(context.Background(), cmd.db)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer dbPool.Close()

	store := db.NewStore(dbPool)
	tx := db.NewTransactor(cmd.log, dbPool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hash := hasher.NewBcrypt()
	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()

	userBus := user.NewService(
		cmd.log,
		store,
		uuidGen,
		clk,
		hash,
		tx,
		eventbus.NewNop(),
	)

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("parsing email: %w", err)
	}

	passd, err := password.ParseConfirm(pass, pass)
	if err != nil {
		return fmt.Errorf("parsing password: %w", err)
	}

	nu := user.NewUser{
		Name:     namePck.MustParse(name),
		Email:    *addr,
		Password: passd,
	}

	usr, err := userBus.Create(ctx, nu)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	fmt.Println("user id:", usr.ID())

	return nil
}
