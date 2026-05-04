package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID                 uuid.UUID   `json:"id"`
	HotelID            uuid.UUID   `json:"hotel_id"`
	RoomTypeID         uuid.UUID   `json:"room_type_id"`
	RoomNumber         string      `json:"room_number"`
	Status             RoomStatus  `json:"status"`
	CleaningStatus     CleanStatus `json:"cleaning_status,omitempty"`
	FloorNumber        int         `json:"floor_number,omitempty"`
	Wing               string      `json:"wing,omitempty"`
	FurnitureCondition string      `json:"furniture_condition,omitempty"`
	LastRenovationDate *time.Time  `json:"last_renovation_date,omitempty"`
	SpecialNotes       string      `json:"special_notes,omitempty"`
	RoomTypeDetail     *RoomType   `json:"room_type_detail,omitempty" db:"-"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}
