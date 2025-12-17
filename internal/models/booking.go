package models

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID         uuid.UUID     `json:"id" db:"id"`
	GuestID    *uuid.UUID    `json:"guest_id,omitempty" db:"guest_id"`
	PropertyID *uuid.UUID    `json:"property_id,omitempty" db:"property_id"`
	RoomID     *uuid.UUID    `json:"room_id,omitempty" db:"room_id"`
	CheckIn    time.Time     `json:"check_in" db:"check_in"`
	CheckOut   time.Time     `json:"check_out" db:"check_out"`
	Nights     int           `json:"nights" db:"nights"`
	TotalPrice float64       `json:"total_price" db:"total_price"`
	Status     BookingStatus `json:"booking_status" db:"booking_status"`
	CreatedAt  time.Time     `json:"created_at" db:"created_at"`
}
