package user_repo_test

import (
	"context"
	"fmt"
	"net/mail"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Housiadas/cerberus/internal/common/dbtest"
	"github.com/Housiadas/cerberus/internal/common/unitest"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/page"
)

func Test_User(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_User")

	sd, err := insertSeedData(db.Core)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, query(db.Core, sd), "query")
	unitest.Run(t, create(db.Core), "create")
	unitest.Run(t, update(db.Core, sd), "update")
	unitest.Run(t, deleteUser(db.Core, sd), "delete")
}

// =============================================================================

func insertSeedData(service dbtest.Service) (unitest.SeedData, error) {
	ctx := context.Background()

	userRoleID := uuid.New()
	usrs, err := user_service.TestSeedUsers(ctx, 2, userRoleID, service.User)
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

	adminRoleID := uuid.New()
	usrs, err = user_service.TestSeedUsers(ctx, 2, adminRoleID, service.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu3 := unitest.User{
		User: usrs[0],
	}

	tu4 := unitest.User{
		User: usrs[1],
	}

	// -------------------------------------------------------------------------

	sd := unitest.SeedData{
		Users:  []unitest.User{tu3, tu4},
		Admins: []unitest.User{tu1, tu2},
	}

	return sd, nil
}

// =============================================================================

func query(service dbtest.Service, sd unitest.SeedData) []unitest.Table {
	usrs := make([]user.User, 0, len(sd.Admins)+len(sd.Users))

	for _, adm := range sd.Admins {
		usrs = append(usrs, adm.User)
	}

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

				resp, err := service.User.Query(ctx, filter, user.DefaultOrderBy, page.MustParse("1", "10"))
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
				resp, err := service.User.QueryByID(ctx, sd.Users[0].ID)
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

func create(service dbtest.Service) []unitest.Table {
	email, _ := mail.ParseAddress("chris@housi.com")

	roleID := uuid.New()
	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: user.User{
				RoleID:     roleID,
				Name:       name.MustParse("Chris Housi"),
				Email:      *email,
				Department: name.MustParseNull("IT0"),
				Enabled:    true,
			},
			ExcFunc: func(ctx context.Context) any {
				nu := user.NewUser{
					Name:       name.MustParse("Chris Housi"),
					Email:      *email,
					RoleID:     roleID,
					Department: name.MustParseNull("IT0"),
					Password:   "123",
				}

				resp, err := service.User.Create(ctx, nu)
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

func update(busDomain dbtest.Service, sd unitest.SeedData) []unitest.Table {
	email, _ := mail.ParseAddress("chris2@housi.com")

	roleID := uuid.New()
	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: user.User{
				ID:         sd.Users[0].ID,
				RoleID:     roleID,
				Name:       name.MustParse("Chris Housi 2"),
				Email:      *email,
				Department: name.MustParseNull("IT0"),
				Enabled:    true,
				CreatedAt:  sd.Users[0].CreatedAt,
			},
			ExcFunc: func(ctx context.Context) any {
				uu := user.UpdateUser{
					RoleID:     roleID,
					Name:       dbtest.NamePointer("Chris Housi 2"),
					Email:      email,
					Department: dbtest.NameNullPointer("IT0"),
					Password:   dbtest.StringPointer("1234"),
				}

				resp, err := busDomain.User.Update(ctx, sd.Users[0].User, uu)
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

func deleteUser(busDomain dbtest.Service, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "user",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.User.Delete(ctx, sd.Users[1].User); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "admin",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.User.Delete(ctx, sd.Admins[1].User); err != nil {
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
