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
	UpdateProperty(property models.Properties) (*models.Properties, error)
	DeleteProperty(id string) error
	ListProperties(city string) ([]models.Properties, error)
	CreateRoomType(roomType models.RoomType) error
	UpdateRoomType(roomType models.RoomType) (*models.RoomType, error)
	DeleteRoomType(id string) error
	ListRoomTypes(propertyID string) ([]models.RoomType, error)
	CreateRoom(room models.Room) error
	UpdateRoom(room models.Room) (*models.Room, error)
	DeleteRoom(id string) error
	ListRooms(propertyID, roomTypeID string) ([]models.Room, error)
	UpsertRoomRates(rates []models.RoomRate) error
	ListRoomRates(roomID string, startDate, endDate string) ([]models.RoomRate, error)
	GetRoomByID(id string) (*models.Room, error)
	GetRoomTypeByID(id string) (*models.RoomType, error)
	GetPropertyPhotoByID(id string) (*models.PropertyPhoto, error)
	GetRoomPhotoByID(id string) (*models.RoomPhoto, error)
	AddPropertyPhoto(photo models.PropertyPhoto) error
	ListPropertyPhotos(propertyID string) ([]models.PropertyPhoto, error)
	DeletePropertyPhoto(id string) error
	AddRoomPhoto(photo models.RoomPhoto) error
	ListRoomPhotos(roomTypeID, roomID string) ([]models.RoomPhoto, error)
	DeleteRoomPhoto(id string) error

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

func (r *propertyRepo) UpdateRoomType(roomType models.RoomType) (*models.RoomType, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	updates := map[string]any{
		"name":        roomType.Name,
		"description": roomType.Description,
		"base_price":  roomType.BasePrice,
		"capacity":    roomType.Capacity,
		"facilities":  roomType.Facilities,
	}
	resp, _, err := config.SupabaseClient.
		From("room_types").
		Update(updates, "", "").
		Eq("id", roomType.ID.String()).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui tipe kamar: %v", err)
	}
	var updated models.RoomType
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *propertyRepo) DeleteRoomType(id string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("room_types").
		Delete("", "").
		Eq("id", id).
		Execute()
	if err != nil {
		return fmt.Errorf("gagal menghapus tipe kamar: %v", err)
	}
	return nil
}

func (r *propertyRepo) ListRoomTypes(propertyID string) ([]models.RoomType, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.
		From("room_types").
		Select("*", "", false)
	if strings.TrimSpace(propertyID) != "" {
		q = q.Eq("property_id", propertyID)
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar tipe kamar: %v", err)
	}
	var roomTypes []models.RoomType
	if err := json.Unmarshal(resp, &roomTypes); err != nil {
		return nil, err
	}
	return roomTypes, nil
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

func (r *propertyRepo) UpdateRoom(room models.Room) (*models.Room, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	updates := map[string]any{
		"room_number":         room.RoomNumber,
		"room_type_id":        room.RoomTypeID,
		"status":              room.Status,
		"housekeeping_status": room.HousekeepingStatus,
	}

	resp, _, err := config.SupabaseClient.
		From("rooms").
		Update(updates, "", "").
		Eq("id", room.ID.String()).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui unit kamar: %v", err)
	}

	var updated models.Room
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *propertyRepo) DeleteRoom(id string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("rooms").
		Delete("", "").
		Eq("id", id).
		Execute()

	if err != nil {
		return fmt.Errorf("gagal menghapus unit kamar: %v", err)
	}

	return nil
}

func (r *propertyRepo) ListRooms(propertyID, roomTypeID string) ([]models.Room, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.
		From("rooms").
		Select("*", "", false)

	if strings.TrimSpace(propertyID) != "" {
		q = q.Eq("property_id", propertyID)
	}
	if strings.TrimSpace(roomTypeID) != "" {
		q = q.Eq("room_type_id", roomTypeID)
	}

	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar unit kamar: %v", err)
	}

	var rooms []models.Room
	if err := json.Unmarshal(resp, &rooms); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *propertyRepo) UpsertRoomRates(rates []models.RoomRate) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	if len(rates) == 0 {
		return fmt.Errorf("rates list cannot be empty")
	}

	_, _, err := config.SupabaseClient.
		From("room_rates").
		Insert(rates, true, "room_id,date", "", ""). // upsert by room_id+date
		Execute()
	if err != nil {
		return fmt.Errorf("gagal menyimpan rate kamar: %v", err)
	}
	return nil
}

func (r *propertyRepo) ListRoomRates(roomID string, startDate, endDate string) ([]models.RoomRate, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.
		From("room_rates").
		Select("*", "", false).
		Eq("room_id", roomID)
	if startDate != "" {
		q = q.Filter("date", "gte", startDate)
	}
	if endDate != "" {
		q = q.Filter("date", "lte", endDate)
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil rate kamar: %v", err)
	}
	var rates []models.RoomRate
	if err := json.Unmarshal(resp, &rates); err != nil {
		return nil, err
	}
	return rates, nil
}

func (r *propertyRepo) GetRoomByID(id string) (*models.Room, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("rooms").
		Select("*", "", false).
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil room: %v", err)
	}
	var room models.Room
	if err := json.Unmarshal(resp, &room); err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *propertyRepo) GetRoomTypeByID(id string) (*models.RoomType, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("room_types").
		Select("*", "", false).
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil room type: %v", err)
	}
	var roomType models.RoomType
	if err := json.Unmarshal(resp, &roomType); err != nil {
		return nil, err
	}
	return &roomType, nil
}

func (r *propertyRepo) AddPropertyPhoto(photo models.PropertyPhoto) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("property_photos").
		Insert(photo, false, "", "", "").
		Execute()
	if err != nil {
		return fmt.Errorf("gagal menambahkan foto property: %v", err)
	}
	return nil
}

func (r *propertyRepo) GetPropertyPhotoByID(id string) (*models.PropertyPhoto, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("property_photos").
		Select("*", "", false).
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil foto property: %v", err)
	}
	var photo models.PropertyPhoto
	if err := json.Unmarshal(resp, &photo); err != nil {
		return nil, err
	}
	return &photo, nil
}

func (r *propertyRepo) ListPropertyPhotos(propertyID string) ([]models.PropertyPhoto, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("property_photos").
		Select("*", "", false).
		Eq("property_id", propertyID).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil foto property: %v", err)
	}
	var photos []models.PropertyPhoto
	if err := json.Unmarshal(resp, &photos); err != nil {
		return nil, err
	}
	return photos, nil
}

func (r *propertyRepo) DeletePropertyPhoto(id string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("property_photos").
		Delete("", "").
		Eq("id", id).
		Execute()
	if err != nil {
		return fmt.Errorf("gagal menghapus foto property: %v", err)
	}
	return nil
}

func (r *propertyRepo) AddRoomPhoto(photo models.RoomPhoto) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("room_photos").
		Insert(photo, false, "", "", "").
		Execute()
	if err != nil {
		return fmt.Errorf("gagal menambahkan foto kamar: %v", err)
	}
	return nil
}

func (r *propertyRepo) GetRoomPhotoByID(id string) (*models.RoomPhoto, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("room_photos").
		Select("*", "", false).
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil foto kamar: %v", err)
	}
	var photo models.RoomPhoto
	if err := json.Unmarshal(resp, &photo); err != nil {
		return nil, err
	}
	return &photo, nil
}

func (r *propertyRepo) ListRoomPhotos(roomTypeID, roomID string) ([]models.RoomPhoto, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.
		From("room_photos").
		Select("*", "", false)
	if strings.TrimSpace(roomTypeID) != "" {
		q = q.Eq("room_type_id", roomTypeID)
	}
	if strings.TrimSpace(roomID) != "" {
		q = q.Eq("room_id", roomID)
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil foto kamar: %v", err)
	}
	var photos []models.RoomPhoto
	if err := json.Unmarshal(resp, &photos); err != nil {
		return nil, err
	}
	return photos, nil
}

func (r *propertyRepo) DeleteRoomPhoto(id string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("room_photos").
		Delete("", "").
		Eq("id", id).
		Execute()
	if err != nil {
		return fmt.Errorf("gagal menghapus foto kamar: %v", err)
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

func (r *propertyRepo) UpdateProperty(property models.Properties) (*models.Properties, error) {
	updates := map[string]any{
		"name":                 property.Name,
		"address":              property.Address,
		"city":                 property.City,
		"facilities":           property.Facilities,
		"checkin_time":         property.CheckInTime,
		"checkout_time":        property.CheckOutTime,
		"cancellation_policy":  property.CancellationPolicy,
	}
	resp, _, err := config.SupabaseClient.
		From("properties").
		Update(updates, "", "").
		Eq("id", property.ID.String()).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah property: %v", err)
	}
	var updated models.Properties
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *propertyRepo) DeleteProperty(id string) error {
	_, _, err := config.SupabaseClient.From("properties").Delete("", "").Eq("id", id).Execute()
	if err != nil {
		return fmt.Errorf("gagal menghapus property: %v", err)
	}
	return nil
}

func (r *propertyRepo) ListProperties(city string) ([]models.Properties, error) {
	q := config.SupabaseClient.From("properties").Select("*", "", false)
	if strings.TrimSpace(city) != "" {
		q = q.Filter("city", "ilike", fmt.Sprintf("%%%s%%", strings.TrimSpace(city)))
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar property: %v", err)
	}
	var props []models.Properties
	if err := json.Unmarshal(resp, &props); err != nil {
		return nil, err
	}
	return props, nil
}
