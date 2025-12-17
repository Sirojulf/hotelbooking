package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID                 uuid.UUID          `json:"id" db:"id"`
	PropertyID         *uuid.UUID         `json:"property_id,omitempty" db:"property_id"`
	RoomNumber         string             `json:"room_number" db:"room_number"`
	RoomTypeID         *uuid.UUID         `json:"room_type_id,omitempty" db:"room_type_id"`
	RoomTypeDetail     *RoomType          `json:"room_type_detail,omitempty" db:"-"`
	Status             RoomStatus         `json:"status" db:"status"`
	HousekeepingStatus HousekeepingStatus `json:"housekeeping_status" db:"housekeeping_status"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at"`
}
