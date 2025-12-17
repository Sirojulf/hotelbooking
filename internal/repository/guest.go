package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

type GuestRepo interface {
	CreateProfile(profile models.Guest) error
	GetGuestByID(id string) (*models.Guest, error)
}

type guestRepo struct{}

func NewGuestRepo() GuestRepo {
	return &guestRepo{}
}

func (r *guestRepo) CreateProfile(profile models.Guest) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("guests").Insert(profile, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("gagal menyisipkan profil tamu ke db: %v", err)
	}
	return nil
}

func (r *guestRepo) GetGuestByID(id string) (*models.Guest, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	resp, _, err := config.SupabaseClient.
		From("guests").
		Select("*", "", false).
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil profil tamu: %v", err)
	}

	var guest models.Guest
	if err := json.Unmarshal(resp, &guest); err != nil {
		return nil, fmt.Errorf("gagal decode profil tamu: %v", err)
	}

	return &guest, nil
}
