package models

import "github.com/google/uuid"

type Guest struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name,omitempty" db:"last_name"`
	Email       string    `json:"email,omitempty" db:"email"`
	Phone       string    `json:"phone" db:"phone"`
	GuestType   GuestType `json:"guest_type" db:"guest_type"`
	Gender      Gender    `json:"gender" db:"gender"`
	VIPStatus   VIPStatus `json:"vip_status" db:"vip_status"`
	Address     string    `json:"address" db:"address"`
	City        string    `json:"city" db:"city"`
	PostalCode  string    `json:"postal_code" db:"postal_code"`
	State       string    `json:"state" db:"state"`
	Country     string    `json:"country" db:"country"`
	Nationality string    `json:"nationality" db:"nationality"`
}
