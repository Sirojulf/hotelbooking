package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"strings"
)

type HotelRepo interface {
	CreateHotel(hotel models.Hotel) error
	GetHotelByID(id string) (*models.Hotel, error)
	ListHotels(search string) ([]models.Hotel, error)
	UpdateHotel(hotel models.Hotel) (*models.Hotel, error)
	DeleteHotel(id string) error

	CreateRoomType(roomType models.RoomType) error
	GetRoomTypeByID(id string) (*models.RoomType, error)
	ListRoomTypes(hotelID string) ([]models.RoomType, error)
	UpdateRoomType(roomType models.RoomType) (*models.RoomType, error)
	DeleteRoomType(id string) error

	CreateRoom(room models.Room) error
	GetRoomByID(id string) (*models.Room, error)
	ListRooms(hotelID, roomTypeID string) ([]models.Room, error)
	UpdateRoom(room models.Room) (*models.Room, error)
	DeleteRoom(id string) error
}

type hotelRepo struct{}

func NewHotelRepo() HotelRepo {
	return &hotelRepo{}
}

// ─── Hotel ────────────────────────────────────────────────────────────────────

func (r *hotelRepo) CreateHotel(hotel models.Hotel) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("hotels").Insert(hotel, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("gagal membuat hotel: %v", err)
	}
	return nil
}

func (r *hotelRepo) GetHotelByID(id string) (*models.Hotel, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.From("hotels").Select("*", "", false).Eq("id", id).Single().Execute()
	if err != nil {
		return nil, fmt.Errorf("hotel tidak ditemukan: %v", err)
	}
	var hotel models.Hotel
	if err := json.Unmarshal(resp, &hotel); err != nil {
		return nil, err
	}
	return &hotel, nil
}

func (r *hotelRepo) ListHotels(search string) ([]models.Hotel, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.From("hotels").Select("*", "", false)
	if strings.TrimSpace(search) != "" {
		q = q.Filter("address", "ilike", fmt.Sprintf("%%%s%%", strings.TrimSpace(search)))
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar hotel: %v", err)
	}
	var hotels []models.Hotel
	if err := json.Unmarshal(resp, &hotels); err != nil {
		return nil, err
	}
	return hotels, nil
}

func (r *hotelRepo) UpdateHotel(hotel models.Hotel) (*models.Hotel, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	updates := map[string]any{
		"name":           hotel.Name,
		"address":        hotel.Address,
		"status":         hotel.Status,
		"check_in_time":  hotel.CheckInTime,
		"check_out_time": hotel.CheckOutTime,
	}
	if hotel.Code != "" {
		updates["code"] = hotel.Code
	}
	if hotel.ImageURL != "" {
		updates["image_url"] = hotel.ImageURL
	}
	resp, _, err := config.SupabaseClient.From("hotels").Update(updates, "", "").Eq("id", hotel.ID.String()).Single().Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui hotel: %v", err)
	}
	var updated models.Hotel
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *hotelRepo) DeleteHotel(id string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("hotels").Delete("", "").Eq("id", id).Execute()
	if err != nil {
		return fmt.Errorf("gagal menghapus hotel: %v", err)
	}
	return nil
}

// ─── RoomType ─────────────────────────────────────────────────────────────────

func (r *hotelRepo) CreateRoomType(roomType models.RoomType) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("room_types").Insert(roomType, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("gagal membuat tipe kamar: %v", err)
	}
	return nil
}

func (r *hotelRepo) GetRoomTypeByID(id string) (*models.RoomType, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.From("room_types").Select("*", "", false).Eq("id", id).Single().Execute()
	if err != nil {
		return nil, fmt.Errorf("tipe kamar tidak ditemukan: %v", err)
	}
	var rt models.RoomType
	if err := json.Unmarshal(resp, &rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *hotelRepo) ListRoomTypes(hotelID string) ([]models.RoomType, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.From("room_types").Select("*", "", false)
	if strings.TrimSpace(hotelID) != "" {
		q = q.Eq("hotel_id", hotelID)
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil tipe kamar: %v", err)
	}
	var rts []models.RoomType
	if err := json.Unmarshal(resp, &rts); err != nil {
		return nil, err
	}
	return rts, nil
}

func (r *hotelRepo) UpdateRoomType(rt models.RoomType) (*models.RoomType, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	updates := map[string]any{
		"name":            rt.Name,
		"price_per_night": rt.PricePerNight,
		"capacity":        rt.Capacity,
		"description":     rt.Description,
	}
	if rt.Amenities != nil {
		updates["amenities"] = rt.Amenities
	}
	if rt.Images != nil {
		updates["images"] = rt.Images
	}
	if rt.BedType != "" {
		updates["bed_type"] = rt.BedType
	}
	if rt.ViewType != "" {
		updates["view_type"] = rt.ViewType
	}
	if rt.SizeSqm > 0 {
		updates["size_sqm"] = rt.SizeSqm
	}
	if rt.BedCount > 0 {
		updates["bed_count"] = rt.BedCount
	}
	updates["smoking_allowed"] = rt.SmokingAllowed
	resp, _, err := config.SupabaseClient.From("room_types").Update(updates, "", "").Eq("id", rt.ID.String()).Single().Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui tipe kamar: %v", err)
	}
	var updated models.RoomType
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *hotelRepo) DeleteRoomType(id string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("room_types").Delete("", "").Eq("id", id).Execute()
	if err != nil {
		return fmt.Errorf("gagal menghapus tipe kamar: %v", err)
	}
	return nil
}

// ─── Room ─────────────────────────────────────────────────────────────────────

func (r *hotelRepo) CreateRoom(room models.Room) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("rooms").Insert(room, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("gagal membuat kamar: %v", err)
	}
	return nil
}

func (r *hotelRepo) GetRoomByID(id string) (*models.Room, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.From("rooms").Select("*", "", false).Eq("id", id).Single().Execute()
	if err != nil {
		return nil, fmt.Errorf("kamar tidak ditemukan: %v", err)
	}
	var room models.Room
	if err := json.Unmarshal(resp, &room); err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *hotelRepo) ListRooms(hotelID, roomTypeID string) ([]models.Room, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.From("rooms").Select("*", "", false)
	if strings.TrimSpace(hotelID) != "" {
		q = q.Eq("hotel_id", hotelID)
	}
	if strings.TrimSpace(roomTypeID) != "" {
		q = q.Eq("room_type_id", roomTypeID)
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar kamar: %v", err)
	}
	var rooms []models.Room
	if err := json.Unmarshal(resp, &rooms); err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *hotelRepo) UpdateRoom(room models.Room) (*models.Room, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	updates := map[string]any{}
	if room.RoomNumber != "" {
		updates["room_number"] = room.RoomNumber
	}
	if room.Status != "" {
		updates["status"] = room.Status
	}
	if room.CleaningStatus != "" {
		updates["cleaning_status"] = room.CleaningStatus
	}
	if room.FloorNumber > 0 {
		updates["floor_number"] = room.FloorNumber
	}
	if room.Wing != "" {
		updates["wing"] = room.Wing
	}
	if room.FurnitureCondition != "" {
		updates["furniture_condition"] = room.FurnitureCondition
	}
	if room.SpecialNotes != "" {
		updates["special_notes"] = room.SpecialNotes
	}
	if len(updates) == 0 {
		return nil, fmt.Errorf("tidak ada field yang diperbarui")
	}
	resp, _, err := config.SupabaseClient.From("rooms").Update(updates, "", "").Eq("id", room.ID.String()).Single().Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui kamar: %v", err)
	}
	var updated models.Room
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *hotelRepo) DeleteRoom(id string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("rooms").Delete("", "").Eq("id", id).Execute()
	if err != nil {
		return fmt.Errorf("gagal menghapus kamar: %v", err)
	}
	return nil
}
