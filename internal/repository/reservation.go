package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"

	"github.com/supabase-community/postgrest-go"
)

type ReservationRepo interface {
	CreateReservation(res models.Reservation) error
	GetReservationByID(id string) (*models.Reservation, error)
	GetReservationsByGuestID(guestID string) ([]models.Reservation, error)
	ListReservations(hotelID, status, startDate, endDate string) ([]models.Reservation, error)
	UpdateReservationPaymentStatus(id string, status models.PaymentStatus) (*models.Reservation, error)
	UpdateReservation(id string, updates map[string]any) (*models.Reservation, error)
	CheckAvailability(roomID, checkIn, checkOut string) (bool, error)
	ListOverlappingReservations(roomID, checkIn, checkOut string) ([]models.Reservation, error)
	DeleteReservation(id string) error
}

type reservationRepo struct{}

func NewReservationRepo() ReservationRepo {
	return &reservationRepo{}
}

func (r *reservationRepo) CreateReservation(res models.Reservation) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	available, err := r.CheckAvailability(res.RoomID.String(), res.CheckInDate.Format("2006-01-02"), res.CheckOutDate.Format("2006-01-02"))
	if err != nil {
		return err
	}
	if !available {
		return fmt.Errorf("kamar tidak tersedia pada tanggal tersebut")
	}
	_, _, err = config.SupabaseClient.From("reservations").Insert(res, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("gagal membuat reservasi: %v", err)
	}
	return nil
}

func (r *reservationRepo) GetReservationByID(id string) (*models.Reservation, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.From("reservations").Select("*", "", false).Eq("id", id).Single().Execute()
	if err != nil {
		return nil, fmt.Errorf("reservasi tidak ditemukan: %v", err)
	}
	var res models.Reservation
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *reservationRepo) GetReservationsByGuestID(guestID string) ([]models.Reservation, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("reservations").
		Select("*, hotels(name, address)", "", false).
		Eq("guest_id", guestID).
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil reservasi: %v", err)
	}
	var reservations []models.Reservation
	if err := json.Unmarshal(resp, &reservations); err != nil {
		return nil, err
	}
	return reservations, nil
}

func (r *reservationRepo) ListReservations(hotelID, status, startDate, endDate string) ([]models.Reservation, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.From("reservations").Select("*", "", false)
	if hotelID != "" {
		q = q.Eq("hotel_id", hotelID)
	}
	if status != "" {
		q = q.Eq("payment_status", status)
	}
	if startDate != "" {
		q = q.Filter("check_out_date", "gt", startDate)
	}
	if endDate != "" {
		q = q.Filter("check_in_date", "lt", endDate)
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar reservasi: %v", err)
	}
	var reservations []models.Reservation
	if err := json.Unmarshal(resp, &reservations); err != nil {
		return nil, err
	}
	return reservations, nil
}

func (r *reservationRepo) UpdateReservationPaymentStatus(id string, status models.PaymentStatus) (*models.Reservation, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("reservations").
		Update(map[string]any{"payment_status": status}, "", "").
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui status pembayaran: %v", err)
	}
	var res models.Reservation
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *reservationRepo) UpdateReservation(id string, updates map[string]any) (*models.Reservation, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("reservations").
		Update(updates, "", "").
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui reservasi: %v", err)
	}
	var res models.Reservation
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *reservationRepo) CheckAvailability(roomID, checkIn, checkOut string) (bool, error) {
	if config.SupabaseClient == nil {
		return false, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("reservations").
		Select("id", "", false).
		Eq("room_id", roomID).
		Filter("payment_status", "neq", string(models.PaymentStatusCancelled)).
		Filter("check_in_date", "lt", checkOut).
		Filter("check_out_date", "gt", checkIn).
		Execute()
	if err != nil {
		return false, fmt.Errorf("gagal mengecek ketersediaan: %v", err)
	}
	var items []map[string]any
	if err := json.Unmarshal(resp, &items); err != nil {
		return false, err
	}
	return len(items) == 0, nil
}

func (r *reservationRepo) ListOverlappingReservations(roomID, checkIn, checkOut string) ([]models.Reservation, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("reservations").
		Select("*", "", false).
		Eq("room_id", roomID).
		Filter("payment_status", "neq", string(models.PaymentStatusCancelled)).
		Filter("check_in_date", "lt", checkOut).
		Filter("check_out_date", "gt", checkIn).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil reservasi overlap: %v", err)
	}
	var reservations []models.Reservation
	if err := json.Unmarshal(resp, &reservations); err != nil {
		return nil, err
	}
	return reservations, nil
}

func (r *reservationRepo) DeleteReservation(id string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("reservations").Delete("", "").Eq("id", id).Execute()
	if err != nil {
		return fmt.Errorf("gagal menghapus reservasi: %v", err)
	}
	return nil
}
