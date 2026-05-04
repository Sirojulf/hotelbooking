package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	FullName  string     `json:"full_name"`
	Role      string     `json:"role,omitempty"`
	HotelID   *uuid.UUID `json:"hotel_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
