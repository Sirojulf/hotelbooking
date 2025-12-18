package models

import (
	"time"

	"github.com/google/uuid"
)

// PropertyPhoto menyimpan URL foto untuk hotel
type PropertyPhoto struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	PropertyID *uuid.UUID `json:"property_id,omitempty" db:"property_id"`
	URL        string     `json:"url" db:"url"`
	Caption    string     `json:"caption,omitempty" db:"caption"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// RoomPhoto menyimpan URL foto untuk tipe kamar/kamar
type RoomPhoto struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	PropertyID *uuid.UUID `json:"property_id,omitempty" db:"property_id"`
	RoomTypeID *uuid.UUID `json:"room_type_id,omitempty" db:"room_type_id"`
	RoomID     *uuid.UUID `json:"room_id,omitempty" db:"room_id"`
	URL        string     `json:"url" db:"url"`
	Caption    string     `json:"caption,omitempty" db:"caption"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}
