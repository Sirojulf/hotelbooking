package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Guest struct {
	ID            uuid.UUID       `json:"id"`
	HotelID       uuid.UUID       `json:"hotel_id"`
	FullName      string          `json:"full_name"`
	Email         string          `json:"email"`
	PhoneNumber   string          `json:"phone_number,omitempty"`
	Title         string          `json:"title,omitempty"`
	LoyaltyTier   string          `json:"loyalty_tier,omitempty"`
	LoyaltyPoints int             `json:"loyalty_points"`
	TotalSpend    float64         `json:"total_spend"`
	TotalStays    int             `json:"total_stays"`
	Preferences   json.RawMessage `json:"preferences,omitempty"`
	LastVisitAt   *time.Time      `json:"last_visit_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}
