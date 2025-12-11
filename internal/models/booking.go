package models

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	GuestID        uuid.UUID     `json:"guest_id" db:"guest_id"`
	RoomID         uuid.UUID     `json:"room_id" db:"room_id"`
	CheckInDate    time.Time     `json:"check_in_date" db:"check_in_date"`
	CheckOutDate   time.Time     `json:"check_out_date" db:"check_out_date"`
	TotalPrice     float64       `json:"total_price" db:"total_price"`
	Status         BookingStatus `json:"status" db:"status"`
	PaymentStatus  string        `json:"payment_status" db:"payment_status"`
	SpecialRequest string        `json:"special_request" db:"special_request"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
}
