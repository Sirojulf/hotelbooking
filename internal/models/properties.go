package models

import (
	"time"

	"github.com/google/uuid"
)

type Properties struct {
	ID        uuid.UUID `json:"id" db:"id"`
	HotelCode string    `json:"hotel_code" db:"hotel_code"`
	AuthCode  string    `json:"auth_code" db:"auth_code"`
	Name      string    `json:"name" db:"name"`
	City      string    `json:"city,omitempty" db:"city"`
	Address   string    `json:"address,omitempty" db:"address"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type PropertyDetailResponse struct {
	Property  *Properties `json:"property"`
	RoomTypes []RoomType  `json:"room_types"`
}
