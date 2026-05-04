package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID            uuid.UUID       `json:"id"`
	ReservationID *uuid.UUID      `json:"reservation_id,omitempty"`
	Description   string          `json:"description"`
	Amount        float64         `json:"amount"`
	Type          TransactionType `json:"type,omitempty"`
	CreatedBy     *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}
