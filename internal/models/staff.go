package models

import "github.com/google/uuid"

type StaffProfile struct {
	ID       uuid.UUID `json:"id" db:"id"`
	FullName string    `json:"full_name" db:"full_name"`
	Email    string    `json:"email" db:"email"`
	Role     string    `json:"role" db:"role"`
}
