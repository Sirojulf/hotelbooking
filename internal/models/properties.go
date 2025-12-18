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
	Facilities []string `json:"facilities,omitempty" db:"facilities"`
	CheckInTime string  `json:"checkin_time,omitempty" db:"checkin_time"`
	CheckOutTime string `json:"checkout_time,omitempty" db:"checkout_time"`
	CancellationPolicy string `json:"cancellation_policy,omitempty" db:"cancellation_policy"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type PropertyDetailResponse struct {
	Property  *Properties `json:"property"`
	RoomTypes []RoomType  `json:"room_types"`
}
