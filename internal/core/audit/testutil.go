package audit

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
)

// TestNewAudits is a helper method for testing.
func TestNewAudits(
	numb int,
	actorID uuid.UUID,
	objEntity entity.Entity,
	action string,
) []NewAudit {
	newAudits := make([]NewAudit, numb)

	for i := range numb {
		objID, _ := uuid.NewV7()
		na := NewAudit{
			ObjID:     objID,
			ObjEntity: objEntity,
			ObjName:   name.MustParse(fmt.Sprintf("ObjName%d", i)),
			ActorID:   actorID,
			Action:    action,
			Data:      struct{ Name string }{Name: fmt.Sprintf("Name%d", i)},
			Message:   fmt.Sprintf("Message%d", i),
		}

		newAudits[i] = na
	}

	return newAudits
}

// TestSeedAudits is a helper method for testing.
func TestSeedAudits(
	ctx context.Context,
	n int,
	actorID uuid.UUID,
	objEntity entity.Entity,
	action string,
	api *Service,
) ([]Audit, error) {
	newAudits := TestNewAudits(n, actorID, objEntity, action)

	audits := make([]Audit, len(newAudits))

	for i, na := range newAudits {
		adt, err := api.Create(ctx, na)
		if err != nil {
			return nil, fmt.Errorf("seeding audit: idx: %d : %w", i, err)
		}

		audits[i] = adt
	}

	return audits, nil
}
