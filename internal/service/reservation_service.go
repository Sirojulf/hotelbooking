package service

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"

	json "github.com/goccy/go-json"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// RPC stats counters — aggregate logging untuk diagnose under load
var (
	rpcSuccess     int64
	rpcEmpty       int64
	rpcUnparseable int64
	rpcErrorCode   int64
	rpcEmptyResv   int64
	rpcConflict    int64

	debugSampleErrCode atomic.Value // *string — simpan 1 sample error response
	debugSampleEmpty   atomic.Value // *string — sample empty reservation
	statsOnce          sync.Once
)

func startRPCStatsReporter() {
	statsOnce.Do(func() {
		go func() {
			t := time.NewTicker(10 * time.Second)
			defer t.Stop()
			for range t.C {
				s := atomic.LoadInt64(&rpcSuccess)
				e := atomic.LoadInt64(&rpcEmpty)
				u := atomic.LoadInt64(&rpcUnparseable)
				c := atomic.LoadInt64(&rpcErrorCode)
				er := atomic.LoadInt64(&rpcEmptyResv)
				cf := atomic.LoadInt64(&rpcConflict)
				if s+e+u+c+er+cf == 0 {
					continue
				}
				log.Printf("[RPC STATS] success=%d empty=%d unparseable=%d errcode=%d emptyresv=%d conflict=%d",
					s, e, u, c, er, cf)
				if v := debugSampleErrCode.Load(); v != nil {
					log.Printf("[RPC SAMPLE errcode] %s", *(v.(*string)))
				}
				if v := debugSampleEmpty.Load(); v != nil {
					log.Printf("[RPC SAMPLE emptyresv] %s", *(v.(*string)))
				}
			}
		}()
	})
}

func debugLogRPC(label, body string) {
	startRPCStatsReporter()
	switch {
	case label == "EMPTY":
		atomic.AddInt64(&rpcEmpty, 1)
	case label == "UNPARSEABLE":
		atomic.AddInt64(&rpcUnparseable, 1)
	case strings.HasPrefix(label, "ERROR_CODE"):
		atomic.AddInt64(&rpcErrorCode, 1)
		if debugSampleErrCode.Load() == nil {
			preview := body
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			s := label + " | " + preview
			debugSampleErrCode.Store(&s)
		}
	case label == "EMPTY_RESERVATION":
		atomic.AddInt64(&rpcEmptyResv, 1)
		if debugSampleEmpty.Load() == nil {
			preview := body
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			s := preview
			debugSampleEmpty.Store(&s)
		}
	}
}

func incRPCSuccess() {
	startRPCStatsReporter()
	atomic.AddInt64(&rpcSuccess, 1)
}

func incRPCConflict() {
	startRPCStatsReporter()
	atomic.AddInt64(&rpcConflict, 1)
}

type ReservationQuote struct {
	room          *models.Room
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
		room:          room,
		Available:     available,
		Nights:        nights,
		TotalPrice:    pricePerNight * float64(nights),
		PricePerNight: pricePerNight,
		Currency:      "IDR",
	}, nil
}

func (s *reservationService) CreateReservation(input CreateReservationInput) (*ReservationCreateResult, error) {
	// Retry RPC pada ROOM_CONFLICT — race kecil dari high-concurrency biasanya
	// hilang dalam 30-150ms (window timing cancel sebelumnya yang baru commit).
	// Exclusion constraint di DB level membuat conflict atomic, retry pasti melihat
	// data terbaru. Total worst-case backoff = 1.57 detik.
	const maxAttempts = 6
	backoffMs := []int{20, 50, 100, 200, 400, 800}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		result, err := s.createViaRPC(input)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// RPC belum di-deploy → langsung fallback (tidak perlu retry)
		if isRPCMissingErr(err) {
			return s.createReservationFallback(input)
		}

		// ROOM_CONFLICT → retry dengan backoff
		if isConflictMsg(err) && attempt < maxAttempts-1 {
			time.Sleep(time.Duration(backoffMs[attempt]) * time.Millisecond)
			continue
		}

		// Error lain → langsung return
		return nil, err
	}
	return nil, lastErr
}

func isConflictMsg(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "tidak tersedia")
}

func isRPCMissingErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// PostgREST balik "function not found" kalau RPC belum dibuat di Supabase
	return strings.Contains(msg, "PGRST202") ||
		strings.Contains(msg, "function_not_found") ||
		strings.Contains(msg, "Could not find the function")
}

// createViaRPC = atomic create via PostgreSQL function di Supabase (1 round-trip).
// Function `create_reservation_atomic` melakukan: lock room row → check overlap →
// INSERT reservation + transaction → return JSONB. Tanpa race condition.
//
// PERLU DEPLOY SQL ke Supabase dulu. Lihat docs/supabase_rpc.sql.
func (s *reservationService) createViaRPC(input CreateReservationInput) (*ReservationCreateResult, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	// Validasi & ambil quote dulu (room/room_type sudah cached, super murah)
	quote, err := s.quoteFromCache(input.RoomID, input.CheckIn, input.CheckOut)
	if err != nil {
		return nil, err
	}
	room := quote.room
	if input.HotelID != "" && room.HotelID.String() != input.HotelID {
		return nil, fmt.Errorf("hotel_id tidak sesuai dengan kamar")
	}

	resID := uuid.New()
	txID := uuid.New()
	params := map[string]any{
		"p_id":               resID,
		"p_hotel_id":         room.HotelID,
		"p_guest_id":         input.GuestID,
		"p_room_id":          input.RoomID,
		"p_check_in":         input.CheckIn.Format("2006-01-02"),
		"p_check_out":        input.CheckOut.Format("2006-01-02"),
		"p_total_price":      quote.TotalPrice,
		"p_booking_source":   input.BookingSource,
		"p_special_requests": input.SpecialRequests,
		"p_payment_method":   input.PaymentMethod,
		"p_tx_id":            txID,
		"p_tx_description":   fmt.Sprintf("Reservasi kamar %d malam", quote.Nights),
	}

	resp := config.SupabaseClient.Rpc("create_reservation_atomic", "", params)
	if resp == "" {
		// Network error / client error → fallback
		debugLogRPC("EMPTY", "")
		return nil, fmt.Errorf("PGRST202")
	}

	// Parse generic dulu untuk deteksi error response dari PostgREST.
	// Success response = {"reservation":{...},"transaction":{...}}
	// Error response   = {"code":"...","message":"...","details":...,"hint":...}
	var probe struct {
		Code        string              `json:"code"`
		Message     string              `json:"message"`
		Reservation *models.Reservation `json:"reservation"`
		Transaction *models.Transaction `json:"transaction"`
	}
	if err := json.Unmarshal([]byte(resp), &probe); err != nil {
		debugLogRPC("UNPARSEABLE", resp)
		return nil, fmt.Errorf("PGRST202") // unparseable → fallback
	}

	// Error response dari PostgREST punya field "code"
	if probe.Code != "" {
		if probe.Code == "P0001" || strings.Contains(probe.Message, "ROOM_CONFLICT") {
			incRPCConflict()
			return nil, fmt.Errorf("kamar tidak tersedia pada tanggal tersebut")
		}
		// Error lain (PGRST202, permission, schema mismatch, dll) → fallback ke flow lama
		debugLogRPC("ERROR_CODE="+probe.Code, resp)
		return nil, fmt.Errorf("PGRST202")
	}

	// Validasi: reservation harus punya ID valid
	if probe.Reservation == nil || probe.Reservation.ID == uuid.Nil {
		debugLogRPC("EMPTY_RESERVATION", resp)
		return nil, fmt.Errorf("PGRST202") // fallback
	}

	incRPCSuccess()
	return &ReservationCreateResult{
		Reservation: probe.Reservation,
		Transaction: probe.Transaction,
		Quote:       quote,
	}, nil
}

// quoteFromCache = quote tanpa hit Supabase sama sekali (full cached).
func (s *reservationService) quoteFromCache(roomID string, checkIn, checkOut time.Time) (*ReservationQuote, error) {
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
	if !isRoomAvailable(room) {
		return nil, fmt.Errorf("kamar tidak tersedia")
	}
	return &ReservationQuote{
		room:          room,
		Available:     true,
		Nights:        nights,
		TotalPrice:    pricePerNight * float64(nights),
		PricePerNight: pricePerNight,
		Currency:      "IDR",
	}, nil
}

func (s *reservationService) createReservationFallback(input CreateReservationInput) (*ReservationCreateResult, error) {
	// Pre-flight check tetap dipakai → reject mayoritas konflik SEBELUM INSERT.
	// Karena GetRoomByID + GetRoomTypeByID sudah di-cache, hanya CheckAvailability yang hit Supabase.
	quote, err := s.QuoteReservation(input.RoomID, input.CheckIn, input.CheckOut)
	if err != nil {
		return nil, err
	}
	if !quote.Available {
		return nil, fmt.Errorf("kamar tidak tersedia pada tanggal tersebut")
	}

	room := quote.room
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
		CheckInDate:     models.NewDate(input.CheckIn),
		CheckOutDate:    models.NewDate(input.CheckOut),
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

	// Paralelisasi: overlap check + insert transaction berjalan concurrent.
	tx := models.Transaction{
		ID:            uuid.New(),
		ReservationID: &res.ID,
		Description:   fmt.Sprintf("Reservasi kamar %d malam", quote.Nights),
		Amount:        quote.TotalPrice,
		Type:          models.TransactionTypeCharge,
		CreatedAt:     time.Now(),
	}

	var overlaps []models.Reservation
	g := new(errgroup.Group)
	g.Go(func() error {
		o, err := s.repo.ListOverlappingReservations(input.RoomID, input.CheckIn.Format("2006-01-02"), input.CheckOut.Format("2006-01-02"))
		if err != nil {
			return err
		}
		overlaps = o
		return nil
	})
	g.Go(func() error {
		return s.txRepo.CreateTransaction(tx)
	})
	if err := g.Wait(); err != nil {
		// Cleanup: hapus reservation DAN transaction (untuk jaga-jaga kalau salah satunya berhasil)
		_ = s.repo.DeleteReservation(res.ID.String())
		_ = s.txRepo.DeleteByReservationID(res.ID.String())
		return nil, err
	}

	// Verifikasi tidak ada race condition (winner = reservation tertua)
	if len(overlaps) > 1 {
		sort.Slice(overlaps, func(i, j int) bool {
			if overlaps[i].CreatedAt.Equal(overlaps[j].CreatedAt) {
				return overlaps[i].ID.String() < overlaps[j].ID.String()
			}
			return overlaps[i].CreatedAt.Before(overlaps[j].CreatedAt)
		})
		if overlaps[0].ID != res.ID {
			// Race-loser: rollback DUA-DUANYA (reservation + transaction yang sudah ter-insert paralel)
			_ = s.repo.DeleteReservation(res.ID.String())
			_ = s.txRepo.DeleteByReservationID(res.ID.String())
			return nil, fmt.Errorf("kamar baru saja terbooking, silakan coba lagi")
		}
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
	if now.After(res.CheckInDate.Time) {
		return 0
	}
	cutoff := res.CheckInDate.Add(-24 * time.Hour)
	if now.Before(cutoff) {
		return res.TotalPrice
	}
	return res.TotalPrice * 0.5
}
