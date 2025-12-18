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
	UpdateHotel(id, name, address, city string, facilities []string, checkIn, checkOut, cancelPolicy string) (*models.Properties, error)
	DeleteHotel(id string) error
	ListHotels(city string) ([]models.Properties, error)
	GetHotelByID(id string) (*models.Properties, error)
	CreateRoomType(propertyID, name, description string, price float64, capacity int, facilities []string) (*models.RoomType, error)
	UpdateRoomType(id, propertyID, name, description string, price float64, capacity int, facilities []string) (*models.RoomType, error)
	DeleteRoomType(id string) error
	ListRoomTypes(propertyID string) ([]models.RoomType, error)
	CreateRoom(propertyID, roomTypeID, roomNumber string) (*models.Room, error)
	UpdateRoom(id, propertyID, roomTypeID, roomNumber string, status models.RoomStatus, hkStatus models.HousekeepingStatus) (*models.Room, error)
	DeleteRoom(id string) error
	ListRooms(propertyID, roomTypeID string) ([]models.Room, error)
	SetRoomRates(rates []models.RoomRate) error
	GetRoomRates(roomID, startDate, endDate string) ([]models.RoomRate, error)
	GetRoomByID(id string) (*models.Room, error)
	GetRoomTypeByID(id string) (*models.RoomType, error)
	GetPropertyPhotoByID(id string) (*models.PropertyPhoto, error)
	GetRoomPhotoByID(id string) (*models.RoomPhoto, error)
	AddPropertyPhoto(propertyID, url, caption string) error
	ListPropertyPhotos(propertyID string) ([]models.PropertyPhoto, error)
	DeletePropertyPhoto(id string) error
	AddRoomPhoto(propertyID, roomTypeID, roomID, url, caption string) error
	ListRoomPhotos(roomTypeID, roomID string) ([]models.RoomPhoto, error)
	DeleteRoomPhoto(id string) error
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

func (s *inventoryService) UpdateRoomType(id, propertyID, name, description string, price float64, capacity int, facilities []string) (*models.RoomType, error) {
	if name == "" {
		return nil, fmt.Errorf("nama tipe kamar wajib diisi")
	}
	if price <= 0 {
		return nil, fmt.Errorf("harga harus lebih dari 0")
	}
	if capacity <= 0 {
		return nil, fmt.Errorf("kapasitas minimal 1 orang")
	}
	roomTypeID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid room type id")
	}
	var propUUID *uuid.UUID
	if propertyID != "" {
		pid, err := uuid.Parse(propertyID)
		if err != nil {
			return nil, fmt.Errorf("property_id tidak valid")
		}
		propUUID = &pid
	}
	return s.repo.UpdateRoomType(models.RoomType{
		ID:          roomTypeID,
		PropertyID:  propUUID,
		Name:        name,
		Description: description,
		BasePrice:   price,
		Capacity:    capacity,
		Facilities:  facilities,
	})
}

func (s *inventoryService) DeleteRoomType(id string) error {
	return s.repo.DeleteRoomType(id)
}

func (s *inventoryService) ListRoomTypes(propertyID string) ([]models.RoomType, error) {
	return s.repo.ListRoomTypes(propertyID)
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

func (s *inventoryService) UpdateRoom(id, propertyID, roomTypeID, roomNumber string, status models.RoomStatus, hkStatus models.HousekeepingStatus) (*models.Room, error) {
	roomUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid room id")
	}
	var propUUID *uuid.UUID
	if propertyID != "" {
		pid, err := uuid.Parse(propertyID)
		if err != nil {
			return nil, fmt.Errorf("invalid property id")
		}
		propUUID = &pid
	}
	var typeUUID *uuid.UUID
	if roomTypeID != "" {
		tid, err := uuid.Parse(roomTypeID)
		if err != nil {
			return nil, fmt.Errorf("invalid room type id")
		}
		typeUUID = &tid
	}
	return s.repo.UpdateRoom(models.Room{
		ID:                 roomUUID,
		PropertyID:         propUUID,
		RoomTypeID:         typeUUID,
		RoomNumber:         roomNumber,
		Status:             status,
		HousekeepingStatus: hkStatus,
	})
}

func (s *inventoryService) DeleteRoom(id string) error {
	return s.repo.DeleteRoom(id)
}

func (s *inventoryService) ListRooms(propertyID, roomTypeID string) ([]models.Room, error) {
	return s.repo.ListRooms(propertyID, roomTypeID)
}

func (s *inventoryService) SetRoomRates(rates []models.RoomRate) error {
	if len(rates) == 0 {
		return fmt.Errorf("rates tidak boleh kosong")
	}
	return s.repo.UpsertRoomRates(rates)
}

func (s *inventoryService) GetRoomRates(roomID, startDate, endDate string) ([]models.RoomRate, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room_id wajib diisi")
	}
	return s.repo.ListRoomRates(roomID, startDate, endDate)
}

func (s *inventoryService) GetRoomByID(id string) (*models.Room, error) {
	if id == "" {
		return nil, fmt.Errorf("room_id wajib diisi")
	}
	return s.repo.GetRoomByID(id)
}

func (s *inventoryService) GetRoomTypeByID(id string) (*models.RoomType, error) {
	if id == "" {
		return nil, fmt.Errorf("room_type_id wajib diisi")
	}
	return s.repo.GetRoomTypeByID(id)
}

func (s *inventoryService) GetPropertyPhotoByID(id string) (*models.PropertyPhoto, error) {
	if id == "" {
		return nil, fmt.Errorf("property_photo_id wajib diisi")
	}
	return s.repo.GetPropertyPhotoByID(id)
}

func (s *inventoryService) GetRoomPhotoByID(id string) (*models.RoomPhoto, error) {
	if id == "" {
		return nil, fmt.Errorf("room_photo_id wajib diisi")
	}
	return s.repo.GetRoomPhotoByID(id)
}

func (s *inventoryService) AddPropertyPhoto(propertyID, url, caption string) error {
	if url == "" {
		return fmt.Errorf("url foto wajib diisi")
	}
	pid, err := uuid.Parse(propertyID)
	if err != nil {
		return fmt.Errorf("invalid property id")
	}
	return s.repo.AddPropertyPhoto(models.PropertyPhoto{
		ID:         uuid.New(),
		PropertyID: &pid,
		URL:        url,
		Caption:    caption,
		CreatedAt:  time.Now(),
	})
}

func (s *inventoryService) ListPropertyPhotos(propertyID string) ([]models.PropertyPhoto, error) {
	if propertyID == "" {
		return nil, fmt.Errorf("property_id wajib diisi")
	}
	return s.repo.ListPropertyPhotos(propertyID)
}

func (s *inventoryService) DeletePropertyPhoto(id string) error {
	return s.repo.DeletePropertyPhoto(id)
}

func (s *inventoryService) AddRoomPhoto(propertyID, roomTypeID, roomID, url, caption string) error {
	if url == "" {
		return fmt.Errorf("url foto wajib diisi")
	}
	var pid *uuid.UUID
	if propertyID != "" {
		parsed, err := uuid.Parse(propertyID)
		if err != nil {
			return fmt.Errorf("invalid property id")
		}
		pid = &parsed
	}
	var rtID *uuid.UUID
	if roomTypeID != "" {
		parsed, err := uuid.Parse(roomTypeID)
		if err != nil {
			return fmt.Errorf("invalid room_type_id")
		}
		rtID = &parsed
	}
	var rID *uuid.UUID
	if roomID != "" {
		parsed, err := uuid.Parse(roomID)
		if err != nil {
			return fmt.Errorf("invalid room_id")
		}
		rID = &parsed
	}
	return s.repo.AddRoomPhoto(models.RoomPhoto{
		ID:         uuid.New(),
		PropertyID: pid,
		RoomTypeID: rtID,
		RoomID:     rID,
		URL:        url,
		Caption:    caption,
		CreatedAt:  time.Now(),
	})
}

func (s *inventoryService) ListRoomPhotos(roomTypeID, roomID string) ([]models.RoomPhoto, error) {
	return s.repo.ListRoomPhotos(roomTypeID, roomID)
}

func (s *inventoryService) DeleteRoomPhoto(id string) error {
	return s.repo.DeleteRoomPhoto(id)
}

func (s *inventoryService) UpdateHotel(id, name, address, city string, facilities []string, checkIn, checkOut, cancelPolicy string) (*models.Properties, error) {
	if name == "" {
		return nil, fmt.Errorf("nama hotel wajib diisi")
	}
	propID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid property id")
	}
	return s.repo.UpdateProperty(models.Properties{
		ID:                  propID,
		Name:                name,
		Address:             address,
		City:                city,
		Facilities:          facilities,
		CheckInTime:         checkIn,
		CheckOutTime:        checkOut,
		CancellationPolicy:  cancelPolicy,
	})
}

func (s *inventoryService) DeleteHotel(id string) error {
	return s.repo.DeleteProperty(id)
}

func (s *inventoryService) ListHotels(city string) ([]models.Properties, error) {
	return s.repo.ListProperties(city)
}

func (s *inventoryService) GetHotelByID(id string) (*models.Properties, error) {
	if id == "" {
		return nil, fmt.Errorf("property_id wajib diisi")
	}
	return s.repo.GetPropertyByID(id)
}
