package user_repo_test

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/password"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/internal/utils/dbtest"
	"github.com/Housiadas/cerberus/internal/utils/unitest"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func Test_User(t *testing.T) {
	t.Parallel()

	// initialize db
	db := sc.NewDB(t)

	// Initialize logger
	var buf bytes.Buffer
	traceIDFn := func(context.Context) string {
		return ""
	}
	requestIDFn := func(context.Context) string {
		return ""
	}
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	// utils
	hash := hasher.NewBcrypt()
	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()
	userService := user_service.New(log, user_repo.NewStore(log, db), uuidGen, clk, hash)

	// seed
	sd, err := insertSeedData(userService)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

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
		return usrs[i].ID().String() <= usrs[j].ID().String()
	})

	userCmpOpts := []cmp.Option{
		cmp.AllowUnexported(user.User{}),
		cmpopts.EquateApproxTime(time.Second),
	}

	table := []unitest.Table{
		{
			Name:    "all",
			ExpResp: usrs,
			ExcFunc: func(ctx context.Context) any {
				filter := user.QueryFilter{
					Name: dbtest.NamePointer("Name"),
				}

				resp, err := service.Query(ctx, filter, user.GetDefaultOrderBy(), mustParseCursor("", "10"))
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

				return cmp.Diff(gotResp, expResp, userCmpOpts...)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Users[0].User,
			ExcFunc: func(ctx context.Context) any {
				resp, err := service.QueryByID(ctx, sd.Users[0].ID())
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

				return cmp.Diff(gotResp, expResp, userCmpOpts...)
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
			ExpResp: user.New(
				uuid.UUID{},
				name.MustParse("Bill Kennedy"),
				*email,
				nil,
				name.MustParseNull("ITO"),
				true,
				nil,
				time.Time{},
				time.Time{},
				nil,
			),
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

				if err := bcrypt.CompareHashAndPassword(gotResp.PasswordHash(), []byte("123")); err != nil {
					return err.Error()
				}

				expResp := exp.(user.User)

				expResp = user.New(
					gotResp.ID(),
					expResp.Name(),
					expResp.Email(),
					gotResp.PasswordHash(),
					expResp.Department(),
					expResp.Enabled(),
					nil,
					gotResp.CreatedAt(),
					gotResp.UpdatedAt(),
					gotResp.DeletedAt(),
				)

				return cmp.Diff(gotResp, expResp, cmp.AllowUnexported(user.User{}))
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
			ExpResp: user.New(
				sd.Users[0].ID(),
				name.MustParse("Jack Kennedy"),
				*email,
				nil,
				name.MustParseNull("ITO"),
				true,
				nil,
				sd.Users[0].CreatedAt(),
				time.Time{},
				nil,
			),
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

				if err := bcrypt.CompareHashAndPassword(gotResp.PasswordHash(), []byte("1234")); err != nil {
					return err.Error()
				}

				expResp := exp.(user.User)

				expResp = user.New(
					expResp.ID(),
					expResp.Name(),
					expResp.Email(),
					gotResp.PasswordHash(),
					expResp.Department(),
					expResp.Enabled(),
					nil,
					expResp.CreatedAt(),
					gotResp.UpdatedAt(),
					expResp.DeletedAt(),
				)

				return cmp.Diff(gotResp, expResp, cmp.AllowUnexported(user.User{}), cmpopts.EquateApproxTime(time.Second))
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

func mustParseCursor(cursorStr, limitStr string) cursor.Cursor {
	cur, err := cursor.Parse(cursorStr, limitStr)
	if err != nil {
		panic(err)
	}

	return cur
}
