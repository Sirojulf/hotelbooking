package models

type Admin struct {
	ID         string `json:"id" db:"id"`
	PropertyID string `json:"property_id" db:"property_id"`
	Email      string `json:"username" db:"email"`
	Password   string `json:"password" db:"password"`
	IsActive   bool   `json:"is_active" db:"is_active"`
	CreatedAt  string `json:"created_at" db:"created_at"`
}
