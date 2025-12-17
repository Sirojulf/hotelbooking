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
