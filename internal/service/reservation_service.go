package service

import (
	"fmt"
	"sort"
	"time"

	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"

	"github.com/google/uuid"
)

type ReservationQuote struct {
	Available     bool    `json:"available"`
	Nights        int     `json:"nights"`
	TotalPrice    float64 `json:"total_price"`
	PricePerNight float64 `json:"price_per_night"`
	Currency      string  `json:"currency"`
}

type CreateReservationInput struct {
	GuestID         string
	HotelID         string
	RoomID          string
	CheckIn         time.Time
	CheckOut        time.Time
	BookingSource   string
	SpecialRequests string
	PaymentMethod   string
}

type ReservationCreateResult struct {
	Reservation *models.Reservation `json:"reservation"`
	Transaction *models.Transaction `json:"transaction"`
	Quote       *ReservationQuote   `json:"quote,omitempty"`
}

type ReservationService interface {
	QuoteReservation(roomID string, checkIn, checkOut time.Time) (*ReservationQuote, error)
	CreateReservation(input CreateReservationInput) (*ReservationCreateResult, error)
	MarkPaymentPaid(guestID, reservationID, paymentMethod string) (*models.Reservation, *models.Transaction, error)
	CancelReservation(guestID, reservationID string) (*models.Reservation, *models.Transaction, error)
	GetTransactions(guestID, reservationID string) ([]models.Transaction, error)
	ListReservations(hotelID, status string, startDate, endDate time.Time) ([]models.Reservation, error)
	UpdateReservation(reservationID string, updates map[string]any) (*models.Reservation, error)
	GetReservationByID(reservationID string) (*models.Reservation, error)
}

type reservationService struct {
	repo      repository.ReservationRepo
	hotelRepo repository.HotelRepo
	txRepo    repository.TransactionRepo
}

func NewReservationService(repo repository.ReservationRepo, hotelRepo repository.HotelRepo, txRepo repository.TransactionRepo) ReservationService {
	return &reservationService{
		repo:      repo,
		hotelRepo: hotelRepo,
		txRepo:    txRepo,
	}
}

func (s *reservationService) QuoteReservation(roomID string, checkIn, checkOut time.Time) (*ReservationQuote, error) {
	nights, err := validateNights(checkIn, checkOut)
	if err != nil {
		return nil, err
	}
	if roomID == "" {
		return nil, fmt.Errorf("room_id wajib diisi")
	}

	room, err := s.hotelRepo.GetRoomByID(roomID)
	if err != nil {
		return nil, err
	}

	roomType, err := s.hotelRepo.GetRoomTypeByID(room.RoomTypeID.String())
	if err != nil {
		return nil, err
	}

	pricePerNight := roomType.PricePerNight
	if pricePerNight == 0 {
		pricePerNight = roomType.BasePrice
	}
	if pricePerNight == 0 {
		return nil, fmt.Errorf("harga kamar belum dikonfigurasi")
	}

	available := isRoomAvailable(room)
	ok, err := s.repo.CheckAvailability(roomID, checkIn.Format("2006-01-02"), checkOut.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	if !ok {
		available = false
	}

	return &ReservationQuote{
		Available:     available,
		Nights:        nights,
		TotalPrice:    pricePerNight * float64(nights),
		PricePerNight: pricePerNight,
		Currency:      "IDR",
	}, nil
}

func (s *reservationService) CreateReservation(input CreateReservationInput) (*ReservationCreateResult, error) {
	quote, err := s.QuoteReservation(input.RoomID, input.CheckIn, input.CheckOut)
	if err != nil {
		return nil, err
	}
	if !quote.Available {
		return nil, fmt.Errorf("kamar tidak tersedia pada tanggal tersebut")
	}

	room, err := s.hotelRepo.GetRoomByID(input.RoomID)
	if err != nil {
		return nil, err
	}
	if input.HotelID != "" && room.HotelID.String() != input.HotelID {
		return nil, fmt.Errorf("hotel_id tidak sesuai dengan kamar")
	}

	guestUUID, err := uuid.Parse(input.GuestID)
	if err != nil {
		return nil, fmt.Errorf("guest_id tidak valid")
	}
	roomUUID, err := uuid.Parse(input.RoomID)
	if err != nil {
		return nil, fmt.Errorf("room_id tidak valid")
	}

	res := models.Reservation{
		ID:              uuid.New(),
		HotelID:         room.HotelID,
		GuestID:         guestUUID,
		RoomID:          roomUUID,
		CheckInDate:     input.CheckIn,
		CheckOutDate:    input.CheckOut,
		TotalPrice:      quote.TotalPrice,
		PaymentStatus:   models.PaymentStatusPending,
		BookingSource:   models.BookingSource(input.BookingSource),
		SpecialRequests: input.SpecialRequests,
		PaymentMethod:   input.PaymentMethod,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateReservation(res); err != nil {
		return nil, err
	}

	overlaps, err := s.repo.ListOverlappingReservations(input.RoomID, input.CheckIn.Format("2006-01-02"), input.CheckOut.Format("2006-01-02"))
	if err != nil {
		_ = s.repo.DeleteReservation(res.ID.String())
		return nil, err
	}
	if len(overlaps) > 1 {
		sort.Slice(overlaps, func(i, j int) bool {
			if overlaps[i].CreatedAt.Equal(overlaps[j].CreatedAt) {
				return overlaps[i].ID.String() < overlaps[j].ID.String()
			}
			return overlaps[i].CreatedAt.Before(overlaps[j].CreatedAt)
		})
		if overlaps[0].ID != res.ID {
			_ = s.repo.DeleteReservation(res.ID.String())
			return nil, fmt.Errorf("kamar baru saja terbooking, silakan coba lagi")
		}
	}

	tx := models.Transaction{
		ID:            uuid.New(),
		ReservationID: &res.ID,
		Description:   fmt.Sprintf("Reservasi kamar %d malam", quote.Nights),
		Amount:        quote.TotalPrice,
		Type:          models.TransactionTypeCharge,
		CreatedAt:     time.Now(),
	}
	if err := s.txRepo.CreateTransaction(tx); err != nil {
		_ = s.repo.DeleteReservation(res.ID.String())
		return nil, err
	}

	return &ReservationCreateResult{
		Reservation: &res,
		Transaction: &tx,
		Quote:       quote,
	}, nil
}

func (s *reservationService) MarkPaymentPaid(guestID, reservationID, paymentMethod string) (*models.Reservation, *models.Transaction, error) {
	res, err := s.repo.GetReservationByID(reservationID)
	if err != nil {
		return nil, nil, err
	}
	if res.GuestID.String() != guestID {
		return nil, nil, fmt.Errorf("reservasi tidak ditemukan")
	}
	if res.PaymentStatus == models.PaymentStatusCancelled {
		return nil, nil, fmt.Errorf("reservasi sudah dibatalkan")
	}
	if res.PaymentStatus == models.PaymentStatusPaid {
		return res, nil, nil
	}

	updates := map[string]any{"payment_status": models.PaymentStatusPaid}
	if paymentMethod != "" {
		updates["payment_method"] = paymentMethod
	}
	updated, err := s.repo.UpdateReservation(reservationID, updates)
	if err != nil {
		return nil, nil, err
	}

	tx := models.Transaction{
		ID:            uuid.New(),
		ReservationID: &res.ID,
		Description:   "Pembayaran reservasi",
		Amount:        res.TotalPrice,
		Type:          models.TransactionTypePayment,
		CreatedAt:     time.Now(),
	}
	if err := s.txRepo.CreateTransaction(tx); err != nil {
		return updated, nil, err
	}

	return updated, &tx, nil
}

func (s *reservationService) CancelReservation(guestID, reservationID string) (*models.Reservation, *models.Transaction, error) {
	res, err := s.repo.GetReservationByID(reservationID)
	if err != nil {
		return nil, nil, err
	}
	if res.GuestID.String() != guestID {
		return nil, nil, fmt.Errorf("reservasi tidak ditemukan")
	}
	if res.PaymentStatus == models.PaymentStatusCancelled {
		return nil, nil, fmt.Errorf("reservasi sudah dibatalkan")
	}

	updated, err := s.repo.UpdateReservationPaymentStatus(reservationID, models.PaymentStatusCancelled)
	if err != nil {
		return nil, nil, err
	}

	if res.PaymentStatus != models.PaymentStatusPaid {
		return updated, nil, nil
	}

	refundAmount := calcRefund(res, time.Now())
	if refundAmount == 0 {
		return updated, nil, nil
	}

	tx := models.Transaction{
		ID:            uuid.New(),
		ReservationID: &res.ID,
		Description:   "Refund pembatalan reservasi",
		Amount:        refundAmount,
		Type:          models.TransactionTypeRefund,
		CreatedAt:     time.Now(),
	}
	if err := s.txRepo.CreateTransaction(tx); err != nil {
		return updated, nil, err
	}

	return updated, &tx, nil
}

func (s *reservationService) GetTransactions(guestID, reservationID string) ([]models.Transaction, error) {
	res, err := s.repo.GetReservationByID(reservationID)
	if err != nil {
		return nil, err
	}
	if res.GuestID.String() != guestID {
		return nil, fmt.Errorf("reservasi tidak ditemukan")
	}
	return s.txRepo.GetTransactionsByReservationID(reservationID)
}

func (s *reservationService) ListReservations(hotelID, status string, startDate, endDate time.Time) ([]models.Reservation, error) {
	var start, end string
	if !startDate.IsZero() {
		start = startDate.Format("2006-01-02")
	}
	if !endDate.IsZero() {
		end = endDate.AddDate(0, 0, 1).Format("2006-01-02")
	}
	return s.repo.ListReservations(hotelID, status, start, end)
}

func (s *reservationService) UpdateReservation(reservationID string, updates map[string]any) (*models.Reservation, error) {
	if reservationID == "" {
		return nil, fmt.Errorf("reservation_id wajib diisi")
	}
	return s.repo.UpdateReservation(reservationID, updates)
}

func (s *reservationService) GetReservationByID(reservationID string) (*models.Reservation, error) {
	return s.repo.GetReservationByID(reservationID)
}

func validateNights(checkIn, checkOut time.Time) (int, error) {
	if checkIn.After(checkOut) {
		return 0, fmt.Errorf("tanggal check-in tidak boleh setelah check-out")
	}
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		return 0, fmt.Errorf("durasi inap minimal 1 malam")
	}
	return nights, nil
}

func isRoomAvailable(room *models.Room) bool {
	if room == nil {
		return false
	}
	if room.Status != models.RoomStatusAvailable {
		return false
	}
	switch room.CleaningStatus {
	case models.CleanStatusClean, models.CleanStatusInspected:
		return true
	default:
		return false
	}
}

func calcRefund(res *models.Reservation, now time.Time) float64 {
	if res == nil {
		return 0
	}
	if now.After(res.CheckInDate) {
		return 0
	}
	cutoff := res.CheckInDate.Add(-24 * time.Hour)
	if now.Before(cutoff) {
		return res.TotalPrice
	}
	return res.TotalPrice * 0.5
}
