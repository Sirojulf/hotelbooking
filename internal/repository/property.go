package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

type PropertyRepo interface {
	GetPropertyByAuth(hotelCode, authCode string) (*models.Properties, error)
	CreateProperty(property models.Properties) error
	CreateRoomType(roomType models.RoomType) error
	CreateRoom(room models.Room) error
}

type propertyRepo struct{}

func NewPropertyRepo() PropertyRepo {
	return &propertyRepo{}
}

func (r *propertyRepo) GetPropertyByAuth(hotelcode, authCode string) (*models.Properties, error) {
	respon, _, err := config.SupabaseClient.
		From("properties").
		Select("*", "", false).
		Eq("hotel_code", hotelcode).
		Eq("auth_code", authCode).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("unauthorized: invalid hotel or auth code")

	}

	var property models.Properties
	if err := json.Unmarshal(respon, &property); err != nil {
		return nil, err
	}

	return &property, nil
}

// CreateProperty
func (r *propertyRepo) CreateProperty(property models.Properties) error {
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
	_, _, err := config.SupabaseClient.
		From("rooms").
		Insert(room, false, "", "", "").
		Execute()

	if err != nil {
		return fmt.Errorf("gagal membuat unit kamar: %v", err)
	}

	return nil
}
