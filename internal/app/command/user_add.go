package command

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	namePck "github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/password"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

// UserAdd adds new users into the database.
func (cmd *Command) UserAdd(name, email, pass string) error {
	if name == "" || email == "" || pass == "" {
		fmt.Println("help: useradd <name> <email> <password>")

		return ErrHelp
	}

	db, err := pgsql.Open(cmd.DB)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hash := hasher.NewBcrypt()
	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()
	userBus := user_service.New(cmd.Log, user_repo.NewStore(cmd.Log, db), uuidGen, clk, hash)

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

	fmt.Println("user id:", usr.ID)

	return nil
}
