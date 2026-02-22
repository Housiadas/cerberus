package audit_service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/web"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	objID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")

	na := audit.NewAudit{
		ObjID:     objID,
		ObjEntity: entity.New("USER"),
		ObjName:   name.MustParse("TestUser"),
		ActorID:   actorID,
		Action:    "create",
		Data:      struct{ Name string }{Name: "Test"},
		Message:   "User created",
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := audit.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("audit.Audit")).Return(nil)

	sut := audit_service.New(mLogger, mStorer)
	aud, err := sut.Create(ctx, na)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, aud.ID)
	assert.Equal(t, objID, aud.ObjID)
	assert.Equal(t, actorID, aud.ActorID)
	assert.Equal(t, "create", aud.Action)
	assert.Equal(t, "User created", aud.Message)
	assert.NotZero(t, aud.Timestamp)
	assert.NotEmpty(t, aud.Data)
}

func TestService_Create_MarshalError(t *testing.T) {
	ctx := context.Background()

	na := audit.NewAudit{
		ObjID:     uuid.MustParse("11234567-89ab-7def-0123-456789abcdef"),
		ObjEntity: entity.New("USER"),
		ObjName:   name.MustParse("TestUser"),
		ActorID:   uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Action:    "create",
		Data:      make(chan int), // channels cannot be marshaled to JSON
		Message:   "User created",
	}

	mLogger := logger.NewMockLogger(t)
	mStorer := audit.NewMockStorer(t)

	sut := audit_service.New(mLogger, mStorer)
	_, err := sut.Create(ctx, na)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "marshal object")
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	na := audit.NewAudit{
		ObjID:     uuid.MustParse("11234567-89ab-7def-0123-456789abcdef"),
		ObjEntity: entity.New("USER"),
		ObjName:   name.MustParse("TestUser"),
		ActorID:   uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Action:    "create",
		Data:      struct{ Name string }{Name: "Test"},
		Message:   "User created",
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := audit.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("audit.Audit")).Return(errors.New("storer error"))

	sut := audit_service.New(mLogger, mStorer)
	_, err := sut.Create(ctx, na)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storer error")
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := audit.QueryFilter{}
	orderBy := order.By{Field: "timestamp", Direction: "desc"}
	page := web.Page{}

	expectedAudits := []audit.Audit{
		{
			ID:      uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			Action:  "create",
			Message: "Test",
		},
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := audit.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, page).Return(expectedAudits, nil)

	sut := audit_service.New(mLogger, mStorer)
	audits, err := sut.Query(ctx, filter, orderBy, page)

	assert.NoError(t, err)
	assert.Len(t, audits, 1)
	assert.Equal(t, expectedAudits[0].ID, audits[0].ID)
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := audit.QueryFilter{}
	orderBy := order.By{Field: "timestamp", Direction: "desc"}
	page := web.Page{}

	mLogger := logger.NewMockLogger(t)

	mStorer := audit.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, page).Return(nil, errors.New("query error"))

	sut := audit_service.New(mLogger, mStorer)
	_, err := sut.Query(ctx, filter, orderBy, page)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}

func TestService_Count_Successful(t *testing.T) {
	ctx := context.Background()
	filter := audit.QueryFilter{}

	mLogger := logger.NewMockLogger(t)

	mStorer := audit.NewMockStorer(t)
	mStorer.EXPECT().Count(ctx, filter).Return(10, nil)

	sut := audit_service.New(mLogger, mStorer)
	count, err := sut.Count(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, 10, count)
}

func TestService_Count_Error(t *testing.T) {
	ctx := context.Background()
	filter := audit.QueryFilter{}

	mLogger := logger.NewMockLogger(t)

	mStorer := audit.NewMockStorer(t)
	mStorer.EXPECT().Count(ctx, filter).Return(0, errors.New("count error"))

	sut := audit_service.New(mLogger, mStorer)
	_, err := sut.Count(ctx, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count error")
}
