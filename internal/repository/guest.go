package repository

import (
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

type GuestRepo interface {
	CreateProfile(profile models.Guest) error
}

type guestRepo struct{}

func NewGuestRepo() GuestRepo {
	return &guestRepo{}
}

func (r *guestRepo) CreateProfile(profile models.Guest) error {
	_, _, err := config.SupabaseClient.From("guests").Insert(profile, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("gagal menyisipkan profil tamu ke db: %v", err)
	}
	return nil
}
