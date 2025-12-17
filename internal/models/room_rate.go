package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RoomRate struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	RoomID           *uuid.UUID      `json:"room_id,omitempty" db:"room_id"`
	Date             time.Time       `json:"date" db:"date"`
	AvailableRooms   int             `json:"available_rooms" db:"available_rooms"`
	LinearRate       *float64        `json:"linear_rate,omitempty" db:"linear_rate"`
	NonLinearRate    json.RawMessage `json:"non_linear_rate,omitempty" db:"non_linear_rate"`
	MinNights        int             `json:"min_nights" db:"min_nights"`
	MaxNights        int             `json:"max_nights" db:"max_nights"`
	StopSell         bool            `json:"stop_sell" db:"stop_sell"`
	CloseOnArrival   bool            `json:"close_on_arrival" db:"close_on_arrival"`
	CloseOnDeparture bool            `json:"close_on_departure" db:"close_on_departure"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
}
