package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RoomType struct {
	ID             uuid.UUID       `json:"id"`
	HotelID        uuid.UUID       `json:"hotel_id"`
	Name           string          `json:"name"`
	PricePerNight  float64         `json:"price_per_night"`
	Capacity       int             `json:"capacity"`
	Amenities      json.RawMessage `json:"amenities,omitempty"`
	BasePrice      float64         `json:"base_price,omitempty"`
	Description    string          `json:"description,omitempty"`
	SizeSqm        float64         `json:"size_sqm,omitempty"`
	BedType        string          `json:"bed_type,omitempty"`
	BedCount       int             `json:"bed_count,omitempty"`
	ViewType       string          `json:"view_type,omitempty"`
	SmokingAllowed bool            `json:"smoking_allowed"`
	Images         json.RawMessage `json:"images,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
