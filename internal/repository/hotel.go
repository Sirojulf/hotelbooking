package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

type HotelRepo interface {
	CreateHotel(hotel models.Hotel) (*models.Hotel, error)
}

type hotelRepo struct{}

func NewHotelRepo() HotelRepo {
	return &hotelRepo{}
}

func (r *hotelRepo) CreateHotel(hotel models.Hotel) (*models.Hotel, error) {
	// Gunakan "representation" untuk meminta data yang baru dibuat dikembalikan
	res, _, err := config.SupabaseClient.From("hotels").Insert(hotel, false, "", "representation", "").Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal menyisipkan data hotel ke db: %v", err)
	}

	// Unmarshal respons JSON ke dalam struct Hotel
	var createdHotels []models.Hotel
	if err := json.Unmarshal(res, &createdHotels); err != nil {
		return nil, fmt.Errorf("gagal unmarshal data hotel: %v", err)
	}

	if len(createdHotels) == 0 {
		return nil, fmt.Errorf("hotel tidak berhasil dibuat")
	}

	return &createdHotels[0], nil
}
