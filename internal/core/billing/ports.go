package billing

import (
	"time"

	"github.com/google/uuid"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}
