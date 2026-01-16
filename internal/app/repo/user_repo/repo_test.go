package user_repo_test

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"sort"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	"github.com/Housiadas/cerberus/internal/common/dbtest"
	"github.com/Housiadas/cerberus/internal/common/unitest"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/password"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/crypto/bcrypt"
)

func Test_User(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------
	db := dbtest.New(t, "Test_User")

	// -------------------------------------------------------------------------

	// Initialize logger
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", "", "")

	// utils
	hash := hasher.NewBcrypt()
	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()
	userService := user_service.New(log, user_repo.NewStore(log, db), uuidGen, clk, hash)

	// -------------------------------------------------------------------------
	sd, err := insertSeedData(userService)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, queryUser(userService, sd), "query")
	unitest.Run(t, createUser(userService), "create")
	unitest.Run(t, updateUser(userService, sd), "update")
	unitest.Run(t, deleteUser(userService, sd), "delete")
}

// =============================================================================

func insertSeedData(service *user_service.Service) (unitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := user_service.TestSeedUsers(ctx, 2, service)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := unitest.User{
		User: usrs[0],
	}

	tu2 := unitest.User{
		User: usrs[1],
	}

	// -------------------------------------------------------------------------

	sd := unitest.SeedData{
		Users: []unitest.User{tu1, tu2},
	}

	return sd, nil
}

// =============================================================================

func queryUser(service *user_service.Service, sd unitest.SeedData) []unitest.Table {
	usrs := make([]user.User, 0, len(sd.Users))

	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	sort.Slice(usrs, func(i, j int) bool {
		return usrs[i].ID.String() <= usrs[j].ID.String()
	})

	table := []unitest.Table{
		{
			Name:    "all",
			ExpResp: usrs,
			ExcFunc: func(ctx context.Context) any {
				filter := user.QueryFilter{
					Name: dbtest.NamePointer("Name"),
				}

				resp, err := service.Query(ctx, filter, user.DefaultOrderBy, web.PageMustParse("1", "10"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]user.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]user.User)

				for i := range gotResp {
					if gotResp[i].CreatedAt.Format(time.RFC3339) == expResp[i].CreatedAt.Format(time.RFC3339) {
						expResp[i].CreatedAt = gotResp[i].CreatedAt
					}

					if gotResp[i].UpdatedAt.Format(time.RFC3339) == expResp[i].UpdatedAt.Format(time.RFC3339) {
						expResp[i].UpdatedAt = gotResp[i].UpdatedAt
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Users[0].User,
			ExcFunc: func(ctx context.Context) any {
				resp, err := service.QueryByID(ctx, sd.Users[0].ID)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(user.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(user.User)

				if gotResp.CreatedAt.Format(time.RFC3339) == expResp.CreatedAt.Format(time.RFC3339) {
					expResp.CreatedAt = gotResp.CreatedAt
				}

				if gotResp.UpdatedAt.Format(time.RFC3339) == expResp.UpdatedAt.Format(time.RFC3339) {
					expResp.UpdatedAt = gotResp.UpdatedAt
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func createUser(service *user_service.Service) []unitest.Table {
	email, _ := mail.ParseAddress("bill@ardanlabs.com")

	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: user.User{
				Name:       name.MustParse("Bill Kennedy"),
				Email:      *email,
				Department: name.MustParseNull("ITO"),
				Enabled:    true,
			},
			ExcFunc: func(ctx context.Context) any {
				nu := user.NewUser{
					Name:       name.MustParse("Bill Kennedy"),
					Email:      *email,
					Department: name.MustParseNull("ITO"),
					Password:   password.MustParse("123"),
				}

				resp, err := service.Create(ctx, nu)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(user.User)
				if !exists {
					return "error occurred"
				}

				if err := bcrypt.CompareHashAndPassword(gotResp.PasswordHash, []byte("123")); err != nil {
					return err.Error()
				}

				expResp := exp.(user.User)

				expResp.ID = gotResp.ID
				expResp.PasswordHash = gotResp.PasswordHash
				expResp.CreatedAt = gotResp.CreatedAt
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func updateUser(service *user_service.Service, sd unitest.SeedData) []unitest.Table {
	email, _ := mail.ParseAddress("jack@housi.com")

	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: user.User{
				ID:         sd.Users[0].ID,
				Name:       name.MustParse("Jack Kennedy"),
				Email:      *email,
				Department: name.MustParseNull("ITO"),
				Enabled:    true,
				CreatedAt:  sd.Users[0].CreatedAt,
			},
			ExcFunc: func(ctx context.Context) any {
				uu := user.UpdateUser{
					Name:       dbtest.NamePointer("Jack Kennedy"),
					Email:      email,
					Department: dbtest.NameNullPointer("ITO"),
					Password:   dbtest.PasswordPointer("1234"),
				}

				resp, err := service.Update(ctx, sd.Users[0].User, uu)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(user.User)
				if !exists {
					return "error occurred"
				}

				if err := bcrypt.CompareHashAndPassword(gotResp.PasswordHash, []byte("1234")); err != nil {
					return err.Error()
				}

				expResp := exp.(user.User)

				expResp.PasswordHash = gotResp.PasswordHash
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func deleteUser(service *user_service.Service, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "user",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := service.Delete(ctx, sd.Users[1].User); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
