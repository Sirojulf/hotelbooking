package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

type PropertyRepo interface {
	GetPropertyByAuth(hotelCode, authCode string) (*models.Properties, error)
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
