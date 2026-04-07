package user_test

import (
	"context"
	"errors"
	"net/mail"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/user"
	usermocks "github.com/Housiadas/cerberus/internal/core/user/mocks"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/unitest"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/internal/types/password"
	"github.com/Housiadas/cerberus/pkg/cursor"
	loggermocks "github.com/Housiadas/cerberus/pkg/logger/mocks"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Authenticate_Successful(t *testing.T) {
	ctx := context.Background()
	email := unitest.MustParseEmail("john@example.com")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		email,
		[]byte("hashed_password"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, email).Return(existingUser, nil)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)

	mHasher := newMockhasher(t)
	mHasher.EXPECT().Compare(existingUser.PasswordHash(), "secret123").Return(nil)

	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	usr, err := sut.Authenticate(ctx, email, "secret123")

	assert.NoError(t, err)
	assert.Equal(t, existingUser.ID(), usr.ID())
	assert.Equal(t, existingUser.Name(), usr.Name())
	assert.Equal(t, existingUser.Email(), usr.Email())
}

func TestService_Authenticate_UserNotFound(t *testing.T) {
	ctx := context.Background()
	email := unitest.MustParseEmail("unknown@example.com")

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, email).Return(user.User{}, user.ErrNotFound)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)
	mHasher := newMockhasher(t)
	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	_, err := sut.Authenticate(ctx, email, "secret123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestService_Authenticate_WrongPassword(t *testing.T) {
	ctx := context.Background()
	email := unitest.MustParseEmail("john@example.com")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		email,
		[]byte("hashed_password"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, email).Return(existingUser, nil)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)

	mHasher := newMockhasher(t)
	mHasher.EXPECT().Compare(existingUser.PasswordHash(), "wrong_password").Return(errors.New("password mismatch"))

	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	_, err := sut.Authenticate(ctx, email, "wrong_password")

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrAuthenticationFailure)
}

func ptrBool(b bool) *bool { return &b }

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	newUser := user.NewUser{
		Name:       name.MustParse("John Doe"),
		Email:      unitest.MustParseEmail("john@example.com"),
		Password:   password.MustParse("password123"),
		Department: name.MustParseNull("Engineering"),
	}
	expectedUser := user.New(
		mUuid,
		name.MustParse("John Doe"),
		unitest.MustParseEmail("john@example.com"),
		[]byte("password123"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, expectedUser).Return(nil)

	mUuidGen := newMockgenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)

	mClock := newMockclock(t)
	mClock.EXPECT().Now().Return(mTime)

	mHasher := newMockhasher(t)
	mHasher.EXPECT().Hash(newUser.Password.String()).Return(expectedUser.PasswordHash(), nil)

	mTx := newMocktransactor(t)
	mTx.EXPECT().RunInTx(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

	mDispatcher := newMockdispatcher(t)
	mDispatcher.EXPECT().Dispatch(ctx, mock.Anything).Return(nil)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	usr, err := sut.Create(ctx, newUser)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, usr.ID())
	assert.Equal(t, newUser.Name, usr.Name())
	assert.Equal(t, newUser.Email, usr.Email())
	assert.Equal(t, newUser.Department, usr.Department())
	assert.NotZero(t, usr.CreatedAt())
	assert.NotZero(t, usr.UpdatedAt())
}

func TestService_Create_Uuid_Error(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	newUser := user.NewUser{
		Name:       name.MustParse("John Doe"),
		Email:      unitest.MustParseEmail("john@example.com"),
		Password:   password.MustParse("password123"),
		Department: name.MustParseNull("Engineering"),
	}

	mLogger := loggermocks.NewMockLogger(t)
	mStorer := usermocks.NewMockStorer(t)

	mUuidGen := newMockgenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, errors.New("uuid initialization error"))

	mClock := newMockclock(t)
	mHasher := newMockhasher(t)
	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	_, err := sut.Create(ctx, newUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uuid initialization error")
}

func TestService_Create_Hasher_Error(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	newUser := user.NewUser{
		Name:       name.MustParse("John Doe"),
		Email:      unitest.MustParseEmail("john@example.com"),
		Password:   password.MustParse("password123"),
		Department: name.MustParseNull("Engineering"),
	}

	mLogger := loggermocks.NewMockLogger(t)
	mStorer := usermocks.NewMockStorer(t)

	mUuidGen := newMockgenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)

	mClock := newMockclock(t)
	mHasher := newMockhasher(t)
	mHasher.EXPECT().
		Hash(newUser.Password.String()).
		Return(nil, errors.New("hash initialization error"))

	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	_, err := sut.Create(ctx, newUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash initialization error")
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	filter := user.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	email, _ := mail.ParseAddress("john@example.com")
	expectedUsers := []user.User{
		user.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			name.MustParse("John Doe"),
			*email,
			nil,
			name.Null{},
			false,
			nil,
			mTime,
			mTime,
			nil,
		),
	}

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedUsers, nil)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)
	mHasher := newMockhasher(t)
	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	users, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, expectedUsers[0].ID(), users[0].ID())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := user.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query failed"))

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)
	mHasher := newMockhasher(t)
	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	email, _ := mail.ParseAddress("john@example.com")
	expectedUser := user.New(
		userID,
		name.MustParse("John Doe"),
		*email,
		nil,
		name.Null{},
		false,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, userID).Return(expectedUser, nil)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)
	mHasher := newMockhasher(t)
	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	usr, err := sut.QueryByID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID(), usr.ID())
	assert.Equal(t, expectedUser.Name(), usr.Name())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, userID).Return(user.User{}, user.ErrNotFound)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)
	mHasher := newMockhasher(t)
	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	_, err := sut.QueryByID(ctx, userID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestService_QueryByEmail_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	email, _ := mail.ParseAddress("john@example.com")

	expectedUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		*email,
		nil,
		name.Null{},
		false,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, *email).Return(expectedUser, nil)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)
	mHasher := newMockhasher(t)
	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	usr, err := sut.QueryByEmail(ctx, *email)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID(), usr.ID())
	assert.Equal(t, expectedUser.Email(), usr.Email())
}

func TestService_QueryByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	email, _ := mail.ParseAddress("unknown@example.com")

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, *email).Return(user.User{}, user.ErrNotFound)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)
	mHasher := newMockhasher(t)
	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	_, err := sut.QueryByEmail(ctx, *email)

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestService_Update_AllFields(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		unitest.MustParseEmail("john@example.com"),
		[]byte("old_hash"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	newName := name.MustParse("Jane Doe")
	newEmail := unitest.MustParseEmail("jane@example.com")
	newPassword := password.MustParse("newpassword123")
	newDepartment := name.MustParseNull("Marketing")
	uu := user.UpdateUser{
		Name:       &newName,
		Email:      &newEmail,
		Password:   &newPassword,
		Department: &newDepartment,
		Enabled:    ptrBool(false),
	}

	expectedUser := user.New(
		existingUser.ID(),
		newName,
		newEmail,
		[]byte("new_hash"),
		newDepartment,
		false,
		nil,
		mTime,
		updatedTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, expectedUser).Return(nil)

	mUuidGen := newMockgenerator(t)

	mClock := newMockclock(t)
	mClock.EXPECT().Now().Return(updatedTime)

	mHasher := newMockhasher(t)
	mHasher.EXPECT().Hash(newPassword.String()).Return([]byte("new_hash"), nil)

	mTx := newMocktransactor(t)
	mTx.EXPECT().RunInTx(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

	mDispatcher := newMockdispatcher(t)
	mDispatcher.EXPECT().Dispatch(ctx, mock.Anything).Return(nil)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	usr, err := sut.Update(ctx, existingUser, uu)

	assert.NoError(t, err)
	assert.Equal(t, newName, usr.Name())
	assert.Equal(t, newEmail, usr.Email())
	assert.Equal(t, newDepartment, usr.Department())
	assert.Equal(t, false, usr.Enabled())
	assert.Equal(t, updatedTime, usr.UpdatedAt())
}

func TestService_Update_PartialFields(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		unitest.MustParseEmail("john@example.com"),
		[]byte("old_hash"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	newName := name.MustParse("Jane Doe")
	uu := user.UpdateUser{
		Name: &newName,
	}

	expectedUser := user.New(
		existingUser.ID(),
		newName,
		existingUser.Email(),
		existingUser.PasswordHash(),
		existingUser.Department(),
		true,
		nil,
		mTime,
		updatedTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, expectedUser).Return(nil)

	mUuidGen := newMockgenerator(t)

	mClock := newMockclock(t)
	mClock.EXPECT().Now().Return(updatedTime)

	mHasher := newMockhasher(t)

	mTx := newMocktransactor(t)
	mTx.EXPECT().RunInTx(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

	mDispatcher := newMockdispatcher(t)
	mDispatcher.EXPECT().Dispatch(ctx, mock.Anything).Return(nil)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	usr, err := sut.Update(ctx, existingUser, uu)

	assert.NoError(t, err)
	assert.Equal(t, newName, usr.Name())
	assert.Equal(t, existingUser.Email(), usr.Email())
}

func TestService_Update_HasherError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		unitest.MustParseEmail("john@example.com"),
		[]byte("old_hash"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	newPassword := password.MustParse("newpassword123")
	uu := user.UpdateUser{
		Password: &newPassword,
	}

	mLogger := loggermocks.NewMockLogger(t)
	mStorer := usermocks.NewMockStorer(t)
	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)

	mHasher := newMockhasher(t)
	mHasher.EXPECT().Hash(newPassword.String()).Return(nil, errors.New("hash error"))

	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	_, err := sut.Update(ctx, existingUser, uu)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash error")
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		unitest.MustParseEmail("john@example.com"),
		[]byte("old_hash"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	newName := name.MustParse("Jane Doe")
	uu := user.UpdateUser{
		Name: &newName,
	}

	expectedUser := user.New(
		existingUser.ID(),
		newName,
		existingUser.Email(),
		existingUser.PasswordHash(),
		existingUser.Department(),
		true,
		nil,
		mTime,
		updatedTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, expectedUser).Return(errors.New("update failed"))

	mUuidGen := newMockgenerator(t)

	mClock := newMockclock(t)
	mClock.EXPECT().Now().Return(updatedTime)

	mHasher := newMockhasher(t)

	mTx := newMocktransactor(t)
	mTx.EXPECT().RunInTx(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	_, err := sut.Update(ctx, existingUser, uu)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	email, _ := mail.ParseAddress("john@example.com")
	usr := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		*email,
		[]byte("hashed_password"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, usr).Return(nil)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)
	mHasher := newMockhasher(t)

	mTx := newMocktransactor(t)
	mTx.EXPECT().RunInTx(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

	mDispatcher := newMockdispatcher(t)
	mDispatcher.EXPECT().Dispatch(ctx, mock.Anything).Return(nil)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	err := sut.Delete(ctx, usr)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	email, _ := mail.ParseAddress("john@example.com")
	usr := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		*email,
		[]byte("hashed_password"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)

	mStorer := usermocks.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, usr).Return(errors.New("delete failed"))

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)
	mHasher := newMockhasher(t)

	mTx := newMocktransactor(t)
	mTx.EXPECT().RunInTx(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	err := sut.Delete(ctx, usr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

func TestService_VerifyPassword_Successful(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	email := unitest.MustParseEmail("john@example.com")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		email,
		[]byte("hashed_password"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)
	mStorer := usermocks.NewMockStorer(t)
	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)

	mHasher := newMockhasher(t)
	mHasher.EXPECT().Compare(existingUser.PasswordHash(), "Secret123!@#").Return(nil)

	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	err := sut.VerifyPassword(existingUser, "Secret123!@#")

	assert.NoError(t, err)
}

func TestService_VerifyPassword_WrongPassword(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	email := unitest.MustParseEmail("john@example.com")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		email,
		[]byte("hashed_password"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := loggermocks.NewMockLogger(t)
	mStorer := usermocks.NewMockStorer(t)
	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)

	mHasher := newMockhasher(t)
	mHasher.EXPECT().Compare(existingUser.PasswordHash(), "wrongPassword").Return(errors.New("password mismatch"))

	mTx := newMocktransactor(t)
	mDispatcher := newMockdispatcher(t)

	sut := user.NewService(mLogger, mStorer, mUuidGen, mClock, mHasher, mTx, mDispatcher)
	err := sut.VerifyPassword(existingUser, "wrongPassword")

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrAuthenticationFailure)
}
