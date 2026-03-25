package reset_token_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	reset_token2 "github.com/Housiadas/cerberus/internal/core/reset_token"
	"github.com/Housiadas/cerberus/internal/core/reset_token/reset_token_service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	id := uuid.MustParse("11111111-1111-7111-1111-111111111111")
	tokenID := uuid.MustParse("22222222-2222-7222-2222-222222222222")
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(id, nil).Once()
	mUuidGen.EXPECT().Generate().Return(tokenID, nil).Once()

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(now)

	mStorer := reset_token2.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock_any_reset_token()).Return(nil)

	sut := reset_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	tkn, err := sut.Create(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, id, tkn.ID())
	assert.Equal(t, userID, tkn.UserID())
	assert.Equal(t, tokenID.String(), tkn.Token())
	assert.False(t, tkn.Used())
	assert.Equal(t, now, tkn.CreatedAt())
	assert.Equal(t, now.UTC().Add(15*time.Minute), tkn.ExpiresAt())
}

func TestService_Create_UUIDError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(uuid.UUID{}, errors.New("uuid error")).Once()

	mClock := clock.NewMockClock(t)
	mStorer := reset_token2.NewMockStorer(t)

	sut := reset_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.Create(ctx, userID)

	assert.Error(t, err)
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	id := uuid.MustParse("11111111-1111-7111-1111-111111111111")
	tokenID := uuid.MustParse("22222222-2222-7222-2222-222222222222")
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(id, nil).Once()
	mUuidGen.EXPECT().Generate().Return(tokenID, nil).Once()

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(now)

	mStorer := reset_token2.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock_any_reset_token()).Return(errors.New("db error"))

	sut := reset_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.Create(ctx, userID)

	assert.Error(t, err)
}

func TestService_QueryByToken_Successful(t *testing.T) {
	ctx := context.Background()
	tokenStr := "some-token-string"
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	expected := reset_token2.New(
		uuid.New(),
		uuid.New(),
		tokenStr,
		now.Add(15*time.Minute),
		now,
		false,
	)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mStorer := reset_token2.NewMockStorer(t)
	mStorer.EXPECT().QueryByToken(ctx, tokenStr).Return(expected, nil)

	sut := reset_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	tkn, err := sut.QueryByToken(ctx, tokenStr)

	require.NoError(t, err)
	assert.Equal(t, expected.ID(), tkn.ID())
	assert.Equal(t, tokenStr, tkn.Token())
}

func TestService_QueryByToken_NotFound(t *testing.T) {
	ctx := context.Background()
	tokenStr := "nonexistent-token"

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mStorer := reset_token2.NewMockStorer(t)
	mStorer.EXPECT().QueryByToken(ctx, tokenStr).Return(reset_token2.ResetToken{}, errors.New("not found"))

	sut := reset_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.QueryByToken(ctx, tokenStr)

	assert.Error(t, err)
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	tkn := reset_token2.New(
		uuid.New(),
		uuid.New(),
		"some-token",
		now.Add(15*time.Minute),
		now,
		false,
	)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mStorer := reset_token2.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, tkn).Return(nil)

	sut := reset_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Delete(ctx, tkn)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	tkn := reset_token2.New(
		uuid.New(),
		uuid.New(),
		"some-token",
		now.Add(15*time.Minute),
		now,
		false,
	)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mStorer := reset_token2.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, tkn).Return(errors.New("db error"))

	sut := reset_token_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Delete(ctx, tkn)

	assert.Error(t, err)
}

// mock_any_reset_token returns a matcher-compatible reset_token value.
// Since we cannot easily match a dynamic struct, we use the actual value built from known inputs.
func mock_any_reset_token() reset_token2.ResetToken {
	id := uuid.MustParse("11111111-1111-7111-1111-111111111111")
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	tokenID := uuid.MustParse("22222222-2222-7222-2222-222222222222")
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	return reset_token2.New(
		id,
		userID,
		tokenID.String(),
		now.UTC().Add(15*time.Minute),
		now,
		false,
	)
}
