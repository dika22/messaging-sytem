package structs

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	ConcurrencyConfig int       `json:"concurrency_config" db:"concurrency_config"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}