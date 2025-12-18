package models

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	PropertyID *uuid.UUID `json:"property_id,omitempty" db:"property_id"`
	Email      string     `json:"email" db:"email"`
	Role       string     `json:"role,omitempty" db:"role"`
	IsActive   bool       `json:"is_active" db:"is_active"`
	CreatedAt  time.Time  `json:"created_at,omitempty" db:"created_at"`
}
