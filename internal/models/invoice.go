package models

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	BookingID     *uuid.UUID    `json:"booking_id,omitempty" db:"booking_id"`
	InvoiceNumber string        `json:"invoice_number" db:"invoice_number"`
	Amount        float64       `json:"amount" db:"amount"`
	Status        PaymentStatus `json:"status" db:"status"`
	IssuedAt      time.Time     `json:"issued_at" db:"issued_at"`
}
