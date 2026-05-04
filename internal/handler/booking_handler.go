package handler

import (
	"hotelbooking/internal/service"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/supabase-community/gotrue-go/types"
)

type BookingHandler struct {
	svc service.ReservationService
}

func NewBookingHandler(svc service.ReservationService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

// @Summary Check room availability and get price quote
// @Tags Rooms
// @Produce json
// @Param room_id path string true "Room ID"
// @Param check_in query string true "Check-in date (YYYY-MM-DD)"
// @Param check_out query string true "Check-out date (YYYY-MM-DD)"
// @Success 200 {object} service.ReservationQuote
// @Failure 400 {object} map[string]string
// @Router /rooms/{room_id}/availability [get]
func (h *BookingHandler) CheckAvailability(c echo.Context) error {
	roomID := c.Param("room_id")
	checkInStr := c.QueryParam("check_in")
	checkOutStr := c.QueryParam("check_out")
	if roomID == "" || checkInStr == "" || checkOutStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "room_id, check_in, check_out wajib diisi"})
	}
	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "format check_in tidak valid"})
	}
	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "format check_out tidak valid"})
	}
	quote, err := h.svc.QuoteReservation(roomID, checkIn, checkOut)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, quote)
}

type CreateReservationRequest struct {
	HotelID         string `json:"hotel_id"`
	RoomID          string `json:"room_id"`
	CheckIn         string `json:"check_in"`
	CheckOut        string `json:"check_out"`
	BookingSource   string `json:"booking_source"`
	SpecialRequests string `json:"special_requests"`
	PaymentMethod   string `json:"payment_method"`
}

// @Summary Create reservation
// @Tags Guests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateReservationRequest true "Create reservation"
// @Success 201 {object} service.ReservationCreateResult
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /guests/reservations [post]
func (h *BookingHandler) CreateReservation(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "format check_in tidak valid"})
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "format check_out tidak valid"})
	}
	result, err := h.svc.CreateReservation(service.CreateReservationInput{
		GuestID:         user.ID.String(),
		HotelID:         req.HotelID,
		RoomID:          req.RoomID,
		CheckIn:         checkIn,
		CheckOut:        checkOut,
		BookingSource:   req.BookingSource,
		SpecialRequests: req.SpecialRequests,
		PaymentMethod:   req.PaymentMethod,
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, result)
}

type PayReservationRequest struct {
	PaymentMethod string `json:"payment_method"`
}

// @Summary Pay reservation
// @Tags Guests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Reservation ID"
// @Param payload body PayReservationRequest true "Payment details"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Router /guests/reservations/{id}/pay [post]
func (h *BookingHandler) PayReservation(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req PayReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	res, tx, err := h.svc.MarkPaymentPaid(user.ID.String(), c.Param("id"), req.PaymentMethod)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"reservation": res, "transaction": tx})
}

// @Summary Cancel reservation
// @Tags Guests
// @Security BearerAuth
// @Produce json
// @Param id path string true "Reservation ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Router /guests/reservations/{id}/cancel [post]
func (h *BookingHandler) CancelReservation(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	res, tx, err := h.svc.CancelReservation(user.ID.String(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"reservation": res, "transaction": tx})
}

// @Summary Get reservation transactions
// @Tags Guests
// @Security BearerAuth
// @Produce json
// @Param id path string true "Reservation ID"
// @Success 200 {array} models.Transaction
// @Failure 401 {object} map[string]string
// @Router /guests/reservations/{id}/transactions [get]
func (h *BookingHandler) GetTransactions(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	txs, err := h.svc.GetTransactions(user.ID.String(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, txs)
}
