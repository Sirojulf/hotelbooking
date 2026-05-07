package models

import (
	"time"

	"github.com/google/uuid"
)

// ReservationHotel holds hotel data embedded by PostgREST relational joins.
type ReservationHotel struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Reservation struct {
	ID              uuid.UUID         `json:"id"`
	HotelID         uuid.UUID         `json:"hotel_id"`
	GuestID         uuid.UUID         `json:"guest_id"`
	RoomID          uuid.UUID         `json:"room_id"`
	CheckInDate     Date              `json:"check_in_date"`
	CheckOutDate    Date              `json:"check_out_date"`
	TotalPrice      float64           `json:"total_price"`
	PaymentStatus   PaymentStatus     `json:"payment_status"`
	BookingSource   BookingSource     `json:"booking_source,omitempty"`
	SpecialRequests string            `json:"special_requests,omitempty"`
	AINotes         string            `json:"ai_notes,omitempty"`
	CheckedInAt     *time.Time        `json:"checked_in_at,omitempty"`
	CheckedOutAt    *time.Time        `json:"checked_out_at,omitempty"`
	PaymentMethod   string            `json:"payment_method,omitempty"`
	AgentID         *uuid.UUID        `json:"agent_id,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	Hotels          *ReservationHotel `json:"hotels,omitempty"`
}
