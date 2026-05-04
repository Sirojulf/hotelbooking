package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

type GuestRepo interface {
	CreateGuest(guest models.Guest) error
	GetGuestByID(id string) (*models.Guest, error)
	GetGuestByEmail(email, hotelID string) (*models.Guest, error)
	UpdateGuestStats(id string, totalSpend float64, totalStays int) error
}

type guestRepo struct{}

func NewGuestRepo() GuestRepo {
	return &guestRepo{}
}

func (r *guestRepo) CreateGuest(guest models.Guest) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("guests").Insert(guest, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("gagal membuat profil tamu: %v", err)
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
		return nil, fmt.Errorf("tamu tidak ditemukan: %v", err)
	}
	var guest models.Guest
	if err := json.Unmarshal(resp, &guest); err != nil {
		return nil, fmt.Errorf("gagal decode profil tamu: %v", err)
	}
	return &guest, nil
}

func (r *guestRepo) GetGuestByEmail(email, hotelID string) (*models.Guest, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.From("guests").Select("*", "", false).Eq("email", email)
	if hotelID != "" {
		q = q.Eq("hotel_id", hotelID)
	}
	resp, _, err := q.Single().Execute()
	if err != nil {
		return nil, fmt.Errorf("tamu tidak ditemukan: %v", err)
	}
	var guest models.Guest
	if err := json.Unmarshal(resp, &guest); err != nil {
		return nil, err
	}
	return &guest, nil
}

func (r *guestRepo) UpdateGuestStats(id string, totalSpend float64, totalStays int) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("guests").
		Update(map[string]any{
			"total_spend": totalSpend,
			"total_stays": totalStays,
		}, "", "").
		Eq("id", id).
		Execute()
	if err != nil {
		return fmt.Errorf("gagal memperbarui statistik tamu: %v", err)
	}
	return nil
}
