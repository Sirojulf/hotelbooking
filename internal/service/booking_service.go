package service

import (
	"fmt"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"time"

	"github.com/google/uuid"
)

type BookingService interface {
	CreateBooking(guestID, propertyID, roomID string, checkIn, checkOut time.Time) (*models.Booking, error)
}

type bookingService struct {
	repo     repository.BookingRepo
	roomRepo repository.PropertyRepo // Kita mungkin butuh info harga kamar dari sini
}

func NewBookingService(repo repository.BookingRepo) BookingService {
	return &bookingService{repo: repo}
}

func (s *bookingService) CreateBooking(guestID, propertyID, roomID string, checkIn, checkOut time.Time) (*models.Booking, error) {
	if checkIn.After(checkOut) {
		return nil, fmt.Errorf("tanggal check-in tidak boleh setelah check-out")
	}

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		return nil, fmt.Errorf("durasi inap minimal 1 malam")
	}

	available, err := s.repo.CheckAvailability(roomID, checkIn.Format("2006-01-02"), checkOut.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, fmt.Errorf("kamar tidak tersedia pada tanggal tersebut")
	}

	// TODO: Ambil harga riil dari room_rates; sementara gunakan harga dummy.
	pricePerNight := 500000.0
	totalPrice := float64(nights) * pricePerNight

	guestUUID, err := uuid.Parse(guestID)
	if err != nil {
		return nil, fmt.Errorf("invalid guest id")
	}
	propertyUUID, err := uuid.Parse(propertyID)
	if err != nil {
		return nil, fmt.Errorf("invalid property id")
	}
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return nil, fmt.Errorf("invalid room id")
	}

	newBooking := models.Booking{
		ID:         uuid.New(),
		GuestID:    &guestUUID,
		PropertyID: &propertyUUID,
		RoomID:     &roomUUID,
		CheckIn:    checkIn,
		CheckOut:   checkOut,
		Nights:     nights,
		TotalPrice: totalPrice,
		Status:     models.BookingStatusConfirmed,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateBooking(newBooking); err != nil {
		return nil, err
	}

	return &newBooking, nil
}
