package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	BookingID *uuid.UUID    `json:"booking_id,omitempty" db:"booking_id"`
	Amount    float64       `json:"amount" db:"amount"`
	Status    PaymentStatus `json:"status" db:"status"`
	Provider  string        `json:"provider,omitempty" db:"provider"`
	Reference string        `json:"reference,omitempty" db:"reference"`
	PaidAt    *time.Time    `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
}
