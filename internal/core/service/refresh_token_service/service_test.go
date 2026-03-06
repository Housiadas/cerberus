package refresh_token_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
	"github.com/Housiadas/cerberus/internal/core/service/refresh_token_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")
	mTokenID := uuid.MustParse("21234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	ttl := 24 * time.Hour

	mLogger := logger.NewMockLogger(t)

	mStorer := refresh_token.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("refresh_token.RefreshToken")).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mID, nil).Once()
	mUuidGen.EXPECT().Generate().Return(mTokenID, nil).Once()

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(mTime)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	tkn, err := sut.Create(ctx, userID, ttl)

	assert.NoError(t, err)
	assert.Equal(t, mID, tkn.ID())
	assert.Equal(t, userID, tkn.UserID())
	assert.Equal(t, mTokenID.String(), tkn.Token())
	assert.Equal(t, mTime, tkn.CreatedAt())
	assert.Equal(t, mTime.UTC().Add(ttl), tkn.ExpiresAt())
	assert.False(t, tkn.Revoked())
}

func TestService_Create_FirstUuidError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	ttl := 24 * time.Hour

	mLogger := logger.NewMockLogger(t)
	mStorer := refresh_token.NewMockStorer(t)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(uuid.Nil, errors.New("uuid error"))

	mClock := clock.NewMockClock(t)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.Create(ctx, userID, ttl)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uuid error")
}

func TestService_Create_SecondUuidError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")
	ttl := 24 * time.Hour

	mLogger := logger.NewMockLogger(t)
	mStorer := refresh_token.NewMockStorer(t)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mID, nil).Once()
	mUuidGen.EXPECT().Generate().Return(uuid.Nil, errors.New("token uuid error")).Once()

	mClock := clock.NewMockClock(t)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.Create(ctx, userID, ttl)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token uuid error")
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")
	mTokenID := uuid.MustParse("21234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	ttl := 24 * time.Hour

	mLogger := logger.NewMockLogger(t)

	mStorer := refresh_token.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("refresh_token.RefreshToken")).Return(errors.New("storer error"))

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mID, nil).Once()
	mUuidGen.EXPECT().Generate().Return(mTokenID, nil).Once()

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(mTime)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.Create(ctx, userID, ttl)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storer error")
}

func TestService_QueryByToken_Successful(t *testing.T) {
	ctx := context.Background()
	token := "some-token-string"

	expectedToken := refresh_token.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.MustParse("11234567-89ab-7def-0123-456789abcdef"),
		token,
		time.Time{},
		time.Time{},
		false,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := refresh_token.NewMockStorer(t)
	mStorer.EXPECT().QueryByToken(ctx, token).Return(expectedToken, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	tkn, err := sut.QueryByToken(ctx, token)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken.ID(), tkn.ID())
	assert.Equal(t, expectedToken.Token(), tkn.Token())
}

func TestService_QueryByToken_Error(t *testing.T) {
	ctx := context.Background()
	token := "some-token-string"

	mLogger := logger.NewMockLogger(t)

	mStorer := refresh_token.NewMockStorer(t)
	mStorer.EXPECT().QueryByToken(ctx, token).Return(refresh_token.RefreshToken{}, errors.New("not found"))

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.QueryByToken(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestService_Revoke_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	tkn := refresh_token.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.MustParse("11234567-89ab-7def-0123-456789abcdef"),
		"some-token",
		mTime.Add(24*time.Hour),
		mTime,
		false,
	)

	revokedToken := tkn.WithRevoked(true)

	mLogger := logger.NewMockLogger(t)

	mStorer := refresh_token.NewMockStorer(t)
	mStorer.EXPECT().Revoke(ctx, revokedToken).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Revoke(ctx, tkn)

	assert.NoError(t, err)
}

func TestService_Revoke_Error(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	tkn := refresh_token.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.MustParse("11234567-89ab-7def-0123-456789abcdef"),
		"some-token",
		mTime.Add(24*time.Hour),
		mTime,
		false,
	)

	revokedToken := tkn.WithRevoked(true)

	mLogger := logger.NewMockLogger(t)

	mStorer := refresh_token.NewMockStorer(t)
	mStorer.EXPECT().Revoke(ctx, revokedToken).Return(errors.New("revoke error"))

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Revoke(ctx, tkn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "revoke error")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()

	tkn := refresh_token.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.MustParse("11234567-89ab-7def-0123-456789abcdef"),
		"some-token",
		time.Time{},
		time.Time{},
		false,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := refresh_token.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, tkn).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Delete(ctx, tkn)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()

	tkn := refresh_token.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.MustParse("11234567-89ab-7def-0123-456789abcdef"),
		"some-token",
		time.Time{},
		time.Time{},
		false,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := refresh_token.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, tkn).Return(errors.New("delete error"))

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := refresh_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Delete(ctx, tkn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}
