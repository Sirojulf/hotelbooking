package service

import (
	"fmt"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"time"

	"github.com/google/uuid"
)

type InventoryService interface {
	CreateHotel(name, address, city, country, hotelCode string) (*models.Properties, error)
}

type inventoryService struct {
	repo repository.PropertyRepo
}

func NewInventoryService(repo repository.PropertyRepo) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) CreateHotel(name, address, city, country, hotelCode string) (*models.Properties, error) {
	// 1. Validasi Input Sederhana
	if name == "" || hotelCode == "" {
		return nil, fmt.Errorf("nama hotel dan kode hotel wajib diisi")
	}

	// 2. Siapkan Model Data
	newProperty := models.Properties{
		ID:        uuid.New(),
		HotelCode: hotelCode,
		// AuthCode bisa digenerate otomatis atau dikosongkan dulu tergantung logic Anda
		AuthCode:  uuid.New().String()[:8],
		Name:      name,
		Address:   address,
		City:      city,
		Country:   country,
		CreatedAt: time.Now(),
	}

	// 3. Panggil Repository
	err := s.repo.CreateProperty(newProperty)
	if err != nil {
		return nil, err
	}

	return &newProperty, nil
}
