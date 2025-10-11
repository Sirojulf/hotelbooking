package models

import (
	"time"

	"github.com/google/uuid"
)

type Hotel struct {
	ID        uuid.UUID `json:"id" db:"id"`
	HotelCode string    `json:"hotel_code" db:"hotel_code"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address,omitempty" db:"address"`
	City      string    `json:"city,omitempty" db:"city"`
	Country   string    `json:"country,omitempty" db:"country"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
}
