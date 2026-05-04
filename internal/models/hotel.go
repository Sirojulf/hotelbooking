package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Hotel struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	Address      string          `json:"address"`
	Status       string          `json:"status"`
	Code         string          `json:"code,omitempty"`
	ImageURL     string          `json:"image_url,omitempty"`
	Settings     json.RawMessage `json:"settings,omitempty"`
	CheckInTime  string          `json:"check_in_time,omitempty"`
	CheckOutTime string          `json:"check_out_time,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type HotelDetailResponse struct {
	Hotel     *Hotel     `json:"hotel"`
	RoomTypes []RoomType `json:"room_types"`
}
