package service

import (
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
)

// PERBAIKAN 1: Ubah return type interface
type CreateHotelService interface {
	Execute(hotel models.Hotel) (*models.Hotel, error)
}

type createHotelService struct {
	hotelRepo repository.HotelRepo
}

func NewCreateHotelService(hotelRepo repository.HotelRepo) CreateHotelService {
	return &createHotelService{hotelRepo: hotelRepo}
}

// PERBAIKAN 2: Kembalikan data hotel dari repository
func (s *createHotelService) Execute(hotel models.Hotel) (*models.Hotel, error) {
	return s.hotelRepo.CreateHotel(hotel)
}
