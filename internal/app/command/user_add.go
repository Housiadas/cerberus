package command

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	namePck "github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_core"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// UserAdd adds new users into the database.
func (cmd *Command) UserAdd(name, email, password string) error {
	if name == "" || email == "" || password == "" {
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

	userBus := user_core.NewCore(cmd.Log, user_repo.NewStore(cmd.Log, db))

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("parsing email: %w", err)
	}

	nu := user.NewUser{
		Name:     namePck.MustParse(name),
		Email:    *addr,
		Password: password,
		Roles:    []role.Role{role.Admin, role.User},
	}

	usr, err := userBus.Create(ctx, nu)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	fmt.Println("user id:", usr.ID)
	return nil
}
