package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"

	// PENTING: Import library postgrest untuk opsi sorting
	"github.com/supabase-community/postgrest-go"
)

type BookingRepo interface {
	CreateBooking(booking models.Booking) error
	CheckAvailability(roomID string, checkIn, checkOut string) (bool, error)
	GetBookingsByGuestID(guestID string) ([]models.Booking, error)
	GetBookingByID(bookingID string) (*models.Booking, error)
	ListBookings(propertyID, status, startDate, endDate string) ([]models.Booking, error)
	UpdateBookingStatus(bookingID string, status models.BookingStatus, note string, refundAmount float64) (*models.Booking, error)
}

type bookingRepo struct{}

func NewBookingRepo() BookingRepo {
	return &bookingRepo{}
}

func (r *bookingRepo) CreateBooking(booking models.Booking) error {
	// Double-check ketersediaan di layer repo untuk mengurangi race condition sederhana
	available, err := r.CheckAvailability(booking.RoomID.String(), booking.CheckIn.Format("2006-01-02"), booking.CheckOut.Format("2006-01-02"))
	if err != nil {
		return err
	}
	if !available {
		return fmt.Errorf("kamar tidak tersedia pada tanggal tersebut")
	}

	_, _, err = config.SupabaseClient.
		From("bookings").
		Insert(booking, false, "", "", "").
		Execute()

	if err != nil {
		return fmt.Errorf("gagal membuat booking: %v", err)
	}
	return nil
}

func (r *bookingRepo) CheckAvailability(roomID string, checkIn, checkOut string) (bool, error) {
	resp, _, err := config.SupabaseClient.
		From("bookings").
		Select("id, check_in, check_out, booking_status", "", false).
		Eq("room_id", roomID).
		Filter("booking_status", "neq", string(models.BookingStatusCancel)).
		Filter("check_in", "lt", checkOut).
		Filter("check_out", "gt", checkIn).
		Execute()

	if err != nil {
		return false, fmt.Errorf("gagal mengecek ketersediaan kamar: %v", err)
	}

	var bookings []models.Booking
	if err := json.Unmarshal(resp, &bookings); err != nil {
		return false, err
	}

	return len(bookings) == 0, nil
}

func (r *bookingRepo) GetBookingsByGuestID(guestID string) ([]models.Booking, error) {
	// Menggunakan postgrest.OrderOpts dari library yang sudah di-import
	resp, _, err := config.SupabaseClient.
		From("bookings").
		Select("*, properties(name, city)", "", false).
		Eq("guest_id", guestID).
		Order("created_at", &postgrest.OrderOpts{Ascending: false}). // Perubahan di sini
		Execute()

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil history booking: %v", err)
	}

	var bookings []models.Booking
	if err := json.Unmarshal(resp, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *bookingRepo) GetBookingByID(bookingID string) (*models.Booking, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("bookings").
		Select("*", "", false).
		Eq("id", bookingID).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil booking: %v", err)
	}
	var booking models.Booking
	if err := json.Unmarshal(resp, &booking); err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepo) ListBookings(propertyID, status, startDate, endDate string) ([]models.Booking, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.
		From("bookings").
		Select("*", "", false)
	if propertyID != "" {
		q = q.Eq("property_id", propertyID)
	}
	if status != "" {
		q = q.Eq("booking_status", status)
	}
	if startDate != "" {
		q = q.Filter("check_in", "gte", startDate)
	}
	if endDate != "" {
		q = q.Filter("check_out", "lte", endDate)
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar booking: %v", err)
	}
	var bookings []models.Booking
	if err := json.Unmarshal(resp, &bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *bookingRepo) UpdateBookingStatus(bookingID string, status models.BookingStatus, note string, refundAmount float64) (*models.Booking, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	updateData := map[string]any{
		"booking_status": status,
	}
	if note != "" {
		updateData["note"] = note
	}
	if refundAmount != 0 || status == models.BookingStatusCancel {
		updateData["refund_amount"] = refundAmount
	}

	resp, _, err := config.SupabaseClient.
		From("bookings").
		Update(updateData, "", "").
		Eq("id", bookingID).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui status booking: %v", err)
	}

	var booking models.Booking
	if err := json.Unmarshal(resp, &booking); err != nil {
		return nil, err
	}
	return &booking, nil
}
