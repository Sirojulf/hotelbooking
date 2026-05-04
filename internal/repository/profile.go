package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

const profileTable = "profiles"
const profileSelect = "id, email, full_name, role, hotel_id, created_at"

type ProfileRepo interface {
	CreateProfile(profile models.Profile) error
	GetProfileByID(id string) (*models.Profile, error)
	GetProfileByEmail(email string) (*models.Profile, error)
	ListProfiles(hotelID string) ([]models.Profile, error)
	UpdateProfileRole(id, role string) error
	UpdateProfileHotel(id, hotelID string) error
}

type profileRepo struct{}

func NewProfileRepo() ProfileRepo {
	return &profileRepo{}
}

func (r *profileRepo) CreateProfile(profile models.Profile) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From(profileTable).
		Insert(profile, false, "", "", "").
		Execute()
	if err != nil {
		return fmt.Errorf("gagal membuat profil: %v", err)
	}
	return nil
}

func (r *profileRepo) GetProfileByID(id string) (*models.Profile, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From(profileTable).
		Select(profileSelect, "", false).
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("profil tidak ditemukan: %v", err)
	}
	var profile models.Profile
	if err := json.Unmarshal(resp, &profile); err != nil {
		return nil, fmt.Errorf("gagal decode profil: %v", err)
	}
	return &profile, nil
}

func (r *profileRepo) GetProfileByEmail(email string) (*models.Profile, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From(profileTable).
		Select(profileSelect, "", false).
		Eq("email", email).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("profil tidak ditemukan: %v", err)
	}
	var profile models.Profile
	if err := json.Unmarshal(resp, &profile); err != nil {
		return nil, fmt.Errorf("gagal decode profil: %v", err)
	}
	return &profile, nil
}

func (r *profileRepo) ListProfiles(hotelID string) ([]models.Profile, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.From(profileTable).Select(profileSelect, "", false)
	if hotelID != "" {
		q = q.Eq("hotel_id", hotelID)
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar profil: %v", err)
	}
	var profiles []models.Profile
	if err := json.Unmarshal(resp, &profiles); err != nil {
		return nil, fmt.Errorf("gagal decode profil: %v", err)
	}
	return profiles, nil
}

func (r *profileRepo) UpdateProfileRole(id, role string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From(profileTable).
		Update(map[string]any{"role": role}, "", "").
		Eq("id", id).
		Execute()
	if err != nil {
		return fmt.Errorf("gagal memperbarui role: %v", err)
	}
	return nil
}

func (r *profileRepo) UpdateProfileHotel(id, hotelID string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	data := map[string]any{"hotel_id": hotelID}
	if hotelID == "" {
		data["hotel_id"] = nil
	}
	_, _, err := config.SupabaseClient.
		From(profileTable).
		Update(data, "", "").
		Eq("id", id).
		Execute()
	if err != nil {
		return fmt.Errorf("gagal memperbarui hotel profil: %v", err)
	}
	return nil
}
