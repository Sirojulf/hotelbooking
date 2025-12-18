package service

import (
	"fmt"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"time"

	"github.com/google/uuid"
)

type NightlyRate struct {
	Date string  `json:"date"`
	Rate float64 `json:"rate"`
}

type BookingQuote struct {
	Available    bool          `json:"available"`
	Nights       int           `json:"nights"`
	TotalPrice   float64       `json:"total_price"`
	NightlyRates []NightlyRate `json:"nightly_rates"`
	Currency     string        `json:"currency,omitempty"`
}

type BookingCreateResult struct {
	Booking *models.Booking `json:"booking"`
	Payment *models.Payment `json:"payment"`
	Invoice *models.Invoice `json:"invoice"`
	Quote   *BookingQuote   `json:"quote,omitempty"`
}

type BookingService interface {
	QuoteBooking(roomID string, checkIn, checkOut time.Time) (*BookingQuote, error)
	CreateBooking(guestID, propertyID, roomID string, checkIn, checkOut time.Time) (*BookingCreateResult, error)
	MarkPaymentPaid(guestID, bookingID, provider, reference string) (*models.Payment, *models.Invoice, error)
	CancelBooking(guestID, bookingID string, now time.Time) (*models.Booking, *models.Payment, error)
	GetInvoice(guestID, bookingID string) (*models.Invoice, error)
	GetPayment(guestID, bookingID string) (*models.Payment, error)
	ListBookings(propertyID, status string, startDate, endDate time.Time) ([]models.Booking, error)
	UpdateStatus(bookingID string, status models.BookingStatus, note string, refundAmount float64) (*models.Booking, error)
	GetBookingByID(bookingID string) (*models.Booking, error)
}

type bookingService struct {
	repo        repository.BookingRepo
	propRepo    repository.PropertyRepo
	paymentRepo repository.PaymentRepo
}

func NewBookingService(repo repository.BookingRepo, propRepo repository.PropertyRepo, paymentRepo repository.PaymentRepo) BookingService {
	return &bookingService{
		repo:        repo,
		propRepo:    propRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *bookingService) QuoteBooking(roomID string, checkIn, checkOut time.Time) (*BookingQuote, error) {
	nights, err := validateStay(checkIn, checkOut)
	if err != nil {
		return nil, err
	}
	if roomID == "" {
		return nil, fmt.Errorf("room_id wajib diisi")
	}

	room, err := s.propRepo.GetRoomByID(roomID)
	if err != nil {
		return nil, err
	}

	basePrice := 0.0
	if room.RoomTypeID != nil {
		roomType, err := s.propRepo.GetRoomTypeByID(room.RoomTypeID.String())
		if err != nil {
			return nil, err
		}
		basePrice = roomType.BasePrice
	}

	startStr := checkIn.Format("2006-01-02")
	endStr := checkOut.AddDate(0, 0, -1).Format("2006-01-02")
	rates, err := s.propRepo.ListRoomRates(roomID, startStr, endStr)
	if err != nil {
		return nil, err
	}
	rateMap := make(map[string]models.RoomRate, len(rates))
	for _, rate := range rates {
		rateMap[rate.Date.Format("2006-01-02")] = rate
	}

	available := true
	total := 0.0
	nightlyRates := make([]NightlyRate, 0, nights)

	for day := checkIn; day.Before(checkOut); day = day.AddDate(0, 0, 1) {
		dateStr := day.Format("2006-01-02")
		roomRate, hasRate := rateMap[dateStr]
		rate := basePrice

		if hasRate {
			if roomRate.LinearRate != nil {
				rate = *roomRate.LinearRate
			}
			if roomRate.StopSell || roomRate.AvailableRooms <= 0 {
				available = false
			}
			if roomRate.CloseOnArrival && dateStr == startStr {
				available = false
			}
			if roomRate.MinNights > 0 && nights < roomRate.MinNights {
				available = false
			}
			if roomRate.MaxNights > 0 && nights > roomRate.MaxNights {
				available = false
			}
		}

		nightlyRates = append(nightlyRates, NightlyRate{
			Date: dateStr,
			Rate: rate,
		})
		total += rate
	}

	if total == 0 {
		return nil, fmt.Errorf("rate tidak tersedia untuk tanggal tersebut")
	}

	ok, err := s.repo.CheckAvailability(roomID, checkIn.Format("2006-01-02"), checkOut.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	if !ok {
		available = false
	}

	return &BookingQuote{
		Available:    available,
		Nights:       nights,
		TotalPrice:   total,
		NightlyRates: nightlyRates,
		Currency:     "IDR",
	}, nil
}

func (s *bookingService) CreateBooking(guestID, propertyID, roomID string, checkIn, checkOut time.Time) (*BookingCreateResult, error) {
	quote, err := s.QuoteBooking(roomID, checkIn, checkOut)
	if err != nil {
		return nil, err
	}
	if !quote.Available {
		return nil, fmt.Errorf("kamar tidak tersedia pada tanggal tersebut")
	}

	room, err := s.propRepo.GetRoomByID(roomID)
	if err != nil {
		return nil, err
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id wajib diisi")
	}
	if room.PropertyID == nil || room.PropertyID.String() != propertyID {
		return nil, fmt.Errorf("property_id tidak sesuai dengan room")
	}

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
		Nights:     quote.Nights,
		TotalPrice: quote.TotalPrice,
		Status:     models.BookingStatusNew,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateBooking(newBooking); err != nil {
		return nil, err
	}

	payment := models.Payment{
		ID:        uuid.New(),
		BookingID: &newBooking.ID,
		Amount:    newBooking.TotalPrice,
		Status:    models.PaymentStatusPending,
		CreatedAt: time.Now(),
	}
	if err := s.paymentRepo.CreatePayment(payment); err != nil {
		return nil, err
	}

	invoice := models.Invoice{
		ID:            uuid.New(),
		BookingID:     &newBooking.ID,
		InvoiceNumber: buildInvoiceNumber(newBooking.ID, newBooking.CreatedAt),
		Amount:        newBooking.TotalPrice,
		Status:        models.PaymentStatusPending,
		IssuedAt:      time.Now(),
	}
	if err := s.paymentRepo.CreateInvoice(invoice); err != nil {
		return nil, err
	}

	return &BookingCreateResult{
		Booking: &newBooking,
		Payment: &payment,
		Invoice: &invoice,
		Quote:   quote,
	}, nil
}

func (s *bookingService) MarkPaymentPaid(guestID, bookingID, provider, reference string) (*models.Payment, *models.Invoice, error) {
	booking, err := s.repo.GetBookingByID(bookingID)
	if err != nil {
		return nil, nil, err
	}
	if booking.GuestID == nil || booking.GuestID.String() != guestID {
		return nil, nil, fmt.Errorf("booking tidak ditemukan")
	}

	payment, err := s.paymentRepo.UpdatePaymentStatus(bookingID, models.PaymentStatusPaid, provider, reference)
	if err != nil {
		return nil, nil, err
	}

	invoice, err := s.paymentRepo.UpdateInvoiceStatus(bookingID, models.PaymentStatusPaid)
	if err != nil {
		return nil, nil, err
	}

	if booking.Status == models.BookingStatusNew {
		if _, err := s.repo.UpdateBookingStatus(bookingID, models.BookingStatusConfirmed, "", 0); err != nil {
			return nil, nil, err
		}
	}

	return payment, invoice, nil
}

func (s *bookingService) CancelBooking(guestID, bookingID string, now time.Time) (*models.Booking, *models.Payment, error) {
	booking, err := s.repo.GetBookingByID(bookingID)
	if err != nil {
		return nil, nil, err
	}
	if booking.GuestID == nil || booking.GuestID.String() != guestID {
		return nil, nil, fmt.Errorf("booking tidak ditemukan")
	}
	if booking.Status == models.BookingStatusCancel || booking.Status == models.BookingStatusCheckedOut {
		return nil, nil, fmt.Errorf("booking tidak dapat dibatalkan")
	}

	refundAmount := calculateRefund(booking, now)
	updated, err := s.repo.UpdateBookingStatus(bookingID, models.BookingStatusCancel, "cancelled_by_guest", refundAmount)
	if err != nil {
		return nil, nil, err
	}

	payment, err := s.paymentRepo.GetPaymentByBookingID(bookingID)
	if err != nil {
		return updated, nil, nil
	}
	if payment.Status != models.PaymentStatusRefunded {
		payment, err = s.paymentRepo.UpdatePaymentStatus(bookingID, models.PaymentStatusRefunded, payment.Provider, payment.Reference)
		if err != nil {
			return updated, nil, err
		}
		_, _ = s.paymentRepo.UpdateInvoiceStatus(bookingID, models.PaymentStatusRefunded)
	}

	return updated, payment, nil
}

func (s *bookingService) GetInvoice(guestID, bookingID string) (*models.Invoice, error) {
	booking, err := s.repo.GetBookingByID(bookingID)
	if err != nil {
		return nil, err
	}
	if booking.GuestID == nil || booking.GuestID.String() != guestID {
		return nil, fmt.Errorf("booking tidak ditemukan")
	}
	return s.paymentRepo.GetInvoiceByBookingID(bookingID)
}

func (s *bookingService) GetPayment(guestID, bookingID string) (*models.Payment, error) {
	booking, err := s.repo.GetBookingByID(bookingID)
	if err != nil {
		return nil, err
	}
	if booking.GuestID == nil || booking.GuestID.String() != guestID {
		return nil, fmt.Errorf("booking tidak ditemukan")
	}
	return s.paymentRepo.GetPaymentByBookingID(bookingID)
}

func (s *bookingService) ListBookings(propertyID, status string, startDate, endDate time.Time) ([]models.Booking, error) {
	var startStr, endStr string
	if !startDate.IsZero() {
		startStr = startDate.Format("2006-01-02")
	}
	if !endDate.IsZero() {
		endStr = endDate.Format("2006-01-02")
	}
	return s.repo.ListBookings(propertyID, status, startStr, endStr)
}

func (s *bookingService) UpdateStatus(bookingID string, status models.BookingStatus, note string, refundAmount float64) (*models.Booking, error) {
	if bookingID == "" {
		return nil, fmt.Errorf("booking_id wajib diisi")
	}
	return s.repo.UpdateBookingStatus(bookingID, status, note, refundAmount)
}

func (s *bookingService) GetBookingByID(bookingID string) (*models.Booking, error) {
	if bookingID == "" {
		return nil, fmt.Errorf("booking_id wajib diisi")
	}
	return s.repo.GetBookingByID(bookingID)
}

func validateStay(checkIn, checkOut time.Time) (int, error) {
	if checkIn.After(checkOut) {
		return 0, fmt.Errorf("tanggal check-in tidak boleh setelah check-out")
	}
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		return 0, fmt.Errorf("durasi inap minimal 1 malam")
	}
	return nights, nil
}

func calculateRefund(booking *models.Booking, now time.Time) float64 {
	if booking == nil {
		return 0
	}
	if now.After(booking.CheckIn) {
		return 0
	}
	cutoff := booking.CheckIn.Add(-24 * time.Hour)
	if now.Before(cutoff) {
		return booking.TotalPrice
	}
	return booking.TotalPrice * 0.5
}

func buildInvoiceNumber(bookingID uuid.UUID, createdAt time.Time) string {
	date := createdAt.Format("20060102")
	return fmt.Sprintf("INV-%s-%s", date, bookingID.String()[:8])
}
