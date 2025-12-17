package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"strings"
)

type PropertyRepo interface {
	// admin
	GetPropertyByAuth(hotelCode, authCode string) (*models.Properties, error)
	CreateProperty(property models.Properties) error
	CreateRoomType(roomType models.RoomType) error
	CreateRoom(room models.Room) error

	// guest
	SearchProperties(city string) ([]models.Properties, error)
	GetPropertyByID(id string) (*models.Properties, error)
	GetRoomTypesByPropertyID(propertyID string) ([]models.RoomType, error)
}

type propertyRepo struct{}

func NewPropertyRepo() PropertyRepo {
	return &propertyRepo{}
}

func (r *propertyRepo) GetPropertyByAuth(hotelcode, authCode string) (*models.Properties, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	respon, _, err := config.SupabaseClient.
		From("properties").
		Select("*", "", false).
		Eq("hotel_code", hotelcode).
		Eq("auth_code", authCode).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil property (periksa hotel/auth code): %w", err)

	}

	var property models.Properties
	if err := json.Unmarshal(respon, &property); err != nil {
		return nil, err
	}

	return &property, nil
}

// CreateProperty
func (r *propertyRepo) CreateProperty(property models.Properties) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("properties").
		Insert(property, false, "", "", "").
		Execute()

	if err != nil {
		return fmt.Errorf("gagal memebuat property hotel: %v", err)
	}

	return nil
}

func (r *propertyRepo) CreateRoomType(roomType models.RoomType) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("room_types").
		Insert(roomType, false, "", "", "").
		Execute()

	if err != nil {
		return fmt.Errorf("gagal membuat tipe kamar: %v", err)
	}
	return nil
}

func (r *propertyRepo) CreateRoom(room models.Room) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("rooms").
		Insert(room, false, "", "", "").
		Execute()

	if err != nil {
		return fmt.Errorf("gagal membuat unit kamar: %v", err)
	}

	return nil
}

// Implementasi Baru: Mencari properti berdasarkan kota (Case insensitive search logic di Supabase agak tricky, kita pakai Eq dulu atau TextSearch jika dikonfigurasi)
func (r *propertyRepo) SearchProperties(city string) ([]models.Properties, error) {
	trimmedCity := strings.TrimSpace(city)
	if trimmedCity == "" {
		return nil, fmt.Errorf("parameter city tidak boleh kosong")
	}

	// Note: Supabase/Postgrest filter 'ilike' formatnya "ilike.%query%"
	// query := fmt.Sprintf("ilike.%%%s%%", city)

	resp, _, err := config.SupabaseClient.
		From("properties").
		Select("*", "", false).
		Filter("city", "ilike", fmt.Sprintf("%%%s%%", trimmedCity)). // Menggunakan filter ilike untuk pencarian
		Execute()

	if err != nil {
		return nil, fmt.Errorf("gagal mencari properti: %v", err)
	}

	var properties []models.Properties
	if err := json.Unmarshal(resp, &properties); err != nil {
		return nil, err
	}

	return properties, nil
}

func (r *propertyRepo) GetPropertyByID(id string) (*models.Properties, error) {
	resp, _, err := config.SupabaseClient.
		From("properties").
		Select("*", "", false).
		Eq("id", id).
		Single().
		Execute()

	if err != nil {
		return nil, err
	}

	var property models.Properties
	if err := json.Unmarshal(resp, &property); err != nil {
		return nil, err
	}
	return &property, nil
}

func (r *propertyRepo) GetRoomTypesByPropertyID(propertyID string) ([]models.RoomType, error) {
	resp, _, err := config.SupabaseClient.
		From("room_types").
		Select("*", "", false).
		Eq("property_id", propertyID).
		Execute()

	if err != nil {
		return nil, err
	}

	var roomTypes []models.RoomType
	if err := json.Unmarshal(resp, &roomTypes); err != nil {
		return nil, err
	}
	return roomTypes, nil
}
