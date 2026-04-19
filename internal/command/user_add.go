package command

import (
	"context"
	"flag"
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

// UserAddRunner implements Runner for the user-add command.
type UserAddRunner struct {
	cmd      *Command
	name     string
	email    string
	password string
}

func NewUserAddRunner(cmd *Command) *UserAddRunner {
	return &UserAddRunner{cmd: cmd}
}

func (r *UserAddRunner) Name() string        { return UserAdd }
func (r *UserAddRunner) Description() string { return "Add a new user to the database" }

func (r *UserAddRunner) SetupFlags(fs *flag.FlagSet) {
	fs.StringVar(&r.name, "name", "", "user's name (required)")
	fs.StringVar(&r.email, "email", "", "user's email (required)")
	fs.StringVar(&r.password, "password", "", "user's password (required)")
}

func (r *UserAddRunner) Run() error {
	if r.name == "" || r.email == "" || r.password == "" {
		return fmt.Errorf("user-add requires -name, -email, and -password flags")
	}

	return r.cmd.userAdd(r.name, r.email, r.password)
}

// userAdd adds new users into the database.
func (cmd *Command) userAdd(name, email, pass string) error {
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
		nil,
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
