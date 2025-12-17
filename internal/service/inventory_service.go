package service

import (
	"fmt"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"time"

	"github.com/google/uuid"
)

type InventoryService interface {
	CreateHotel(name, address, city, hotelCode string) (*models.Properties, error)
	CreateRoomType(propertyID, name, description string, price float64, capacity int, facilities []string) (*models.RoomType, error)
	CreateRoom(propertyID, roomTypeID, roomNumber string) (*models.Room, error)
}

type inventoryService struct {
	repo repository.PropertyRepo
}

func NewInventoryService(repo repository.PropertyRepo) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) CreateHotel(name, address, city, hotelCode string) (*models.Properties, error) {
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
		CreatedAt: time.Now(),
	}

	// 3. Panggil Repository
	err := s.repo.CreateProperty(newProperty)
	if err != nil {
		return nil, err
	}

	return &newProperty, nil
}

func (s *inventoryService) CreateRoomType(propertyID, name, description string, price float64, capacity int, facilities []string) (*models.RoomType, error) {
	if name == "" {
		return nil, fmt.Errorf("nama tipe kamar wajib diisi")
	}

	if price <= 0 {
		return nil, fmt.Errorf("harga harus lebih dari 0")
	}

	if capacity <= 0 {
		return nil, fmt.Errorf("kapasitas minimal 1 orang")
	}

	propUUID, err := uuid.Parse(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property_id tidak valid")
	}

	newRoomType := models.RoomType{
		ID:          uuid.New(),
		PropertyID:  &propUUID,
		Name:        name,
		Description: description,
		BasePrice:   price,
		Capacity:    capacity,
		Facilities:  facilities,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateRoomType(newRoomType); err != nil {
		return nil, err
	}

	return &newRoomType, nil
}

func (s *inventoryService) CreateRoom(propertyID, roomTypeID, roomNumber string) (*models.Room, error) {
	if roomNumber == "" {
		return nil, fmt.Errorf("nomor kamar wajib diisi")
	}

	propUUID, err := uuid.Parse(propertyID)
	if err != nil {
		return nil, fmt.Errorf("invalid property id")
	}
	typeUUID, err := uuid.Parse(roomTypeID)
	if err != nil {
		return nil, fmt.Errorf("invalid room id")
	}

	newRoom := models.Room{
		ID:                 uuid.New(),
		PropertyID:         &propUUID,
		RoomTypeID:         &typeUUID,
		RoomNumber:         roomNumber,
		RoomTypeDetail:     &models.RoomType{},
		Status:             models.RoomStatusAvailable,
		HousekeepingStatus: models.HousekeepingStatusClean,
		CreatedAt:          time.Now(),
	}

	if err := s.repo.CreateRoom(newRoom); err != nil {
		return nil, err
	}

	return &newRoom, nil
}
