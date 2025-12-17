package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomType struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	PropertyID  *uuid.UUID `json:"property_id,omitempty" db:"property_id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	BasePrice   float64    `json:"base_price" db:"base_price"`
	Capacity    int        `json:"capacity" db:"capacity"`
	Facilities  []string   `json:"facilities" db:"facilities"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}
