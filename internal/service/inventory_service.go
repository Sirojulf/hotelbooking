package service

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"time"

	"github.com/google/uuid"
)

type CreateHotelInput struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	Code         string `json:"code"`
	ImageURL     string `json:"image_url"`
	CheckInTime  string `json:"check_in_time"`
	CheckOutTime string `json:"check_out_time"`
}

type UpdateHotelInput struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	Status       string `json:"status"`
	Code         string `json:"code"`
	ImageURL     string `json:"image_url"`
	CheckInTime  string `json:"check_in_time"`
	CheckOutTime string `json:"check_out_time"`
}

type CreateRoomTypeInput struct {
	HotelID        string          `json:"hotel_id"`
	Name           string          `json:"name"`
	PricePerNight  float64         `json:"price_per_night"`
	Capacity       int             `json:"capacity"`
	Description    string          `json:"description"`
	BedType        string          `json:"bed_type"`
	BedCount       int             `json:"bed_count"`
	ViewType       string          `json:"view_type"`
	SizeSqm        float64         `json:"size_sqm"`
	SmokingAllowed bool            `json:"smoking_allowed"`
	Amenities      json.RawMessage `json:"amenities"`
	Images         json.RawMessage `json:"images"`
}

type UpdateRoomTypeInput struct {
	Name           string          `json:"name"`
	PricePerNight  float64         `json:"price_per_night"`
	Capacity       int             `json:"capacity"`
	Description    string          `json:"description"`
	BedType        string          `json:"bed_type"`
	BedCount       int             `json:"bed_count"`
	ViewType       string          `json:"view_type"`
	SizeSqm        float64         `json:"size_sqm"`
	SmokingAllowed bool            `json:"smoking_allowed"`
	Amenities      json.RawMessage `json:"amenities"`
	Images         json.RawMessage `json:"images"`
}

type CreateRoomInput struct {
	HotelID    string `json:"hotel_id"`
	RoomTypeID string `json:"room_type_id"`
	RoomNumber string `json:"room_number"`
	FloorNumber int   `json:"floor_number"`
	Wing       string `json:"wing"`
}

type UpdateRoomInput struct {
	RoomNumber         string             `json:"room_number"`
	Status             models.RoomStatus  `json:"status"`
	CleaningStatus     models.CleanStatus `json:"cleaning_status"`
	FloorNumber        int                `json:"floor_number"`
	Wing               string             `json:"wing"`
	FurnitureCondition string             `json:"furniture_condition"`
	SpecialNotes       string             `json:"special_notes"`
}

type InventoryService interface {
	CreateHotel(input CreateHotelInput) (*models.Hotel, error)
	UpdateHotel(id string, input UpdateHotelInput) (*models.Hotel, error)
	DeleteHotel(id string) error
	ListHotels(search string) ([]models.Hotel, error)
	GetHotelByID(id string) (*models.Hotel, error)

	CreateRoomType(input CreateRoomTypeInput) (*models.RoomType, error)
	UpdateRoomType(id string, input UpdateRoomTypeInput) (*models.RoomType, error)
	DeleteRoomType(id string) error
	ListRoomTypes(hotelID string) ([]models.RoomType, error)
	GetRoomTypeByID(id string) (*models.RoomType, error)

	CreateRoom(input CreateRoomInput) (*models.Room, error)
	UpdateRoom(id string, input UpdateRoomInput) (*models.Room, error)
	DeleteRoom(id string) error
	ListRooms(hotelID, roomTypeID string) ([]models.Room, error)
	GetRoomByID(id string) (*models.Room, error)
}

type inventoryService struct {
	repo repository.HotelRepo
}

func NewInventoryService(repo repository.HotelRepo) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) CreateHotel(input CreateHotelInput) (*models.Hotel, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("nama hotel wajib diisi")
	}
	hotel := models.Hotel{
		ID:           uuid.New(),
		Name:         input.Name,
		Address:      input.Address,
		Status:       string(models.HotelStatusActive),
		Code:         input.Code,
		ImageURL:     input.ImageURL,
		CheckInTime:  input.CheckInTime,
		CheckOutTime: input.CheckOutTime,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.CreateHotel(hotel); err != nil {
		return nil, err
	}
	return &hotel, nil
}

func (s *inventoryService) UpdateHotel(id string, input UpdateHotelInput) (*models.Hotel, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("nama hotel wajib diisi")
	}
	hotelID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("hotel_id tidak valid")
	}
	status := input.Status
	if status == "" {
		status = string(models.HotelStatusActive)
	}
	return s.repo.UpdateHotel(models.Hotel{
		ID:           hotelID,
		Name:         input.Name,
		Address:      input.Address,
		Status:       status,
		Code:         input.Code,
		ImageURL:     input.ImageURL,
		CheckInTime:  input.CheckInTime,
		CheckOutTime: input.CheckOutTime,
	})
}

func (s *inventoryService) DeleteHotel(id string) error        { return s.repo.DeleteHotel(id) }
func (s *inventoryService) ListHotels(search string) ([]models.Hotel, error) {
	return s.repo.ListHotels(search)
}
func (s *inventoryService) GetHotelByID(id string) (*models.Hotel, error) {
	return s.repo.GetHotelByID(id)
}

func (s *inventoryService) CreateRoomType(input CreateRoomTypeInput) (*models.RoomType, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("nama tipe kamar wajib diisi")
	}
	if input.PricePerNight <= 0 {
		return nil, fmt.Errorf("harga per malam harus lebih dari 0")
	}
	if input.Capacity <= 0 {
		return nil, fmt.Errorf("kapasitas minimal 1")
	}
	hotelUUID, err := uuid.Parse(input.HotelID)
	if err != nil {
		return nil, fmt.Errorf("hotel_id tidak valid")
	}
	rt := models.RoomType{
		ID:             uuid.New(),
		HotelID:        hotelUUID,
		Name:           input.Name,
		PricePerNight:  input.PricePerNight,
		Capacity:       input.Capacity,
		Description:    input.Description,
		BedType:        input.BedType,
		BedCount:       input.BedCount,
		ViewType:       input.ViewType,
		SizeSqm:        input.SizeSqm,
		SmokingAllowed: input.SmokingAllowed,
		Amenities:      input.Amenities,
		Images:         input.Images,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.repo.CreateRoomType(rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (s *inventoryService) UpdateRoomType(id string, input UpdateRoomTypeInput) (*models.RoomType, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("nama tipe kamar wajib diisi")
	}
	rtID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("room_type_id tidak valid")
	}
	return s.repo.UpdateRoomType(models.RoomType{
		ID:             rtID,
		Name:           input.Name,
		PricePerNight:  input.PricePerNight,
		Capacity:       input.Capacity,
		Description:    input.Description,
		BedType:        input.BedType,
		BedCount:       input.BedCount,
		ViewType:       input.ViewType,
		SizeSqm:        input.SizeSqm,
		SmokingAllowed: input.SmokingAllowed,
		Amenities:      input.Amenities,
		Images:         input.Images,
	})
}

func (s *inventoryService) DeleteRoomType(id string) error { return s.repo.DeleteRoomType(id) }
func (s *inventoryService) ListRoomTypes(hotelID string) ([]models.RoomType, error) {
	return s.repo.ListRoomTypes(hotelID)
}
func (s *inventoryService) GetRoomTypeByID(id string) (*models.RoomType, error) {
	return s.repo.GetRoomTypeByID(id)
}

func (s *inventoryService) CreateRoom(input CreateRoomInput) (*models.Room, error) {
	if input.RoomNumber == "" {
		return nil, fmt.Errorf("nomor kamar wajib diisi")
	}
	hotelUUID, err := uuid.Parse(input.HotelID)
	if err != nil {
		return nil, fmt.Errorf("hotel_id tidak valid")
	}
	typeUUID, err := uuid.Parse(input.RoomTypeID)
	if err != nil {
		return nil, fmt.Errorf("room_type_id tidak valid")
	}
	room := models.Room{
		ID:             uuid.New(),
		HotelID:        hotelUUID,
		RoomTypeID:     typeUUID,
		RoomNumber:     input.RoomNumber,
		Status:         models.RoomStatusAvailable,
		CleaningStatus: models.CleanStatusClean,
		FloorNumber:    input.FloorNumber,
		Wing:           input.Wing,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.repo.CreateRoom(room); err != nil {
		return nil, err
	}
	return &room, nil
}

func (s *inventoryService) UpdateRoom(id string, input UpdateRoomInput) (*models.Room, error) {
	roomID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("room_id tidak valid")
	}
	return s.repo.UpdateRoom(models.Room{
		ID:                 roomID,
		RoomNumber:         input.RoomNumber,
		Status:             input.Status,
		CleaningStatus:     input.CleaningStatus,
		FloorNumber:        input.FloorNumber,
		Wing:               input.Wing,
		FurnitureCondition: input.FurnitureCondition,
		SpecialNotes:       input.SpecialNotes,
	})
}

func (s *inventoryService) DeleteRoom(id string) error { return s.repo.DeleteRoom(id) }
func (s *inventoryService) ListRooms(hotelID, roomTypeID string) ([]models.Room, error) {
	return s.repo.ListRooms(hotelID, roomTypeID)
}
func (s *inventoryService) GetRoomByID(id string) (*models.Room, error) {
	return s.repo.GetRoomByID(id)
}
