package plan_repo

import (
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/plan"
	"github.com/google/uuid"
)

type planDB struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Interval    string    `db:"interval"`
	PriceCents  int       `db:"price_cents"`
	Currency    string    `db:"currency"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toPlanDB(p plan.Plan) planDB {
	return planDB{
		ID:          p.ID(),
		Name:        p.Name().String(),
		Description: p.Description(),
		Interval:    p.Interval().String(),
		PriceCents:  p.PriceCents(),
		Currency:    p.Currency().String(),
		IsActive:    p.IsActive(),
		CreatedAt:   p.CreatedAt().UTC(),
		UpdatedAt:   p.UpdatedAt().UTC(),
	}
}

func toPlanDomain(db planDB) (plan.Plan, error) {
	nme, err := name.Parse(db.Name)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("parse name: %w", err)
	}

	interval, err := plan.Parse(db.Interval)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("parse interval: %w", err)
	}

	cur, err := currency.Parse(db.Currency)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("parse currency: %w", err)
	}

	return plan.New(
		db.ID,
		nme,
		db.Description,
		interval,
		db.PriceCents,
		cur,
		db.IsActive,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	), nil
}

func toPlansDomain(dbs []planDB) ([]plan.Plan, error) {
	items := make([]plan.Plan, len(dbs))

	for i, db := range dbs {
		var err error

		items[i], err = toPlanDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to plans domain error: %w", err)
		}
	}

	return items, nil
}
