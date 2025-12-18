package handler

import (
	"hotelbooking/internal/models"
	"hotelbooking/internal/service"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/supabase-community/gotrue-go/types"
)

type BookingHandler struct {
	Svc service.BookingService
}

func NewBookingHandler(svc service.BookingService) *BookingHandler {
	return &BookingHandler{Svc: svc}
}

type AvailabilityResponse struct {
	Available    bool                  `json:"available"`
	Nights       int                   `json:"nights"`
	TotalPrice   float64               `json:"total_price"`
	NightlyRates []service.NightlyRate `json:"nightly_rates"`
	Currency     string                `json:"currency,omitempty"`
}

type PaymentInvoiceResponse struct {
	Payment *models.Payment `json:"payment"`
	Invoice *models.Invoice `json:"invoice"`
}

type BookingCancelResponse struct {
	Booking *models.Booking `json:"booking"`
	Payment *models.Payment `json:"payment"`
}

// GET /api/v1/rooms/:room_id/availability?check_in=YYYY-MM-DD&check_out=YYYY-MM-DD
// @Summary Check room availability
// @Tags Rooms
// @Produce json
// @Param room_id path string true "Room ID"
// @Param check_in query string true "Check-in date (YYYY-MM-DD)"
// @Param check_out query string true "Check-out date (YYYY-MM-DD)"
// @Success 200 {object} AvailabilityResponse
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /rooms/{room_id}/availability [get]
func (h *BookingHandler) CheckAvailability(c echo.Context) error {
	roomID := c.Param("room_id")
	checkInStr := c.QueryParam("check_in")
	checkOutStr := c.QueryParam("check_out")
	if roomID == "" || checkInStr == "" || checkOutStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "room_id, check_in, and check_out are required"})
	}
	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid check_in"})
	}
	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid check_out"})
	}

	quote, err := h.Svc.QuoteBooking(roomID, checkIn, checkOut)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, AvailabilityResponse{
		Available:    quote.Available,
		Nights:       quote.Nights,
		TotalPrice:   quote.TotalPrice,
		NightlyRates: quote.NightlyRates,
		Currency:     quote.Currency,
	})
}

type CreateBookingRequest struct {
	PropertyID string `json:"property_id"`
	RoomID     string `json:"room_id"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
}

// POST /api/v1/guests/bookings
// @Summary Create booking
// @Tags Guests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateBookingRequest true "Create booking"
// @Success 201 {object} service.BookingCreateResult
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /guests/bookings [post]
func (h *BookingHandler) CreateBooking(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	var req CreateBookingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid check_in"})
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid check_out"})
	}

	result, err := h.Svc.CreateBooking(user.ID.String(), req.PropertyID, req.RoomID, checkIn, checkOut)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

type PayBookingRequest struct {
	Provider  string `json:"provider"`
	Reference string `json:"reference"`
}

// POST /api/v1/guests/bookings/:id/pay
// @Summary Pay booking
// @Tags Guests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param payload body PayBookingRequest true "Payment payload"
// @Success 200 {object} PaymentInvoiceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /guests/bookings/{id}/pay [post]
func (h *BookingHandler) PayBooking(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	bookingID := c.Param("id")
	var req PayBookingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	payment, invoice, err := h.Svc.MarkPaymentPaid(user.ID.String(), bookingID, req.Provider, req.Reference)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"payment": payment,
		"invoice": invoice,
	})
}

// POST /api/v1/guests/bookings/:id/cancel
// @Summary Cancel booking
// @Tags Guests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} BookingCancelResponse
// @Failure 401 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /guests/bookings/{id}/cancel [post]
func (h *BookingHandler) CancelBooking(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	bookingID := c.Param("id")

	booking, payment, err := h.Svc.CancelBooking(user.ID.String(), bookingID, time.Now())
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"booking": booking,
		"payment": payment,
	})
}

// GET /api/v1/guests/bookings/:id/invoice
// @Summary Get booking invoice
// @Tags Guests
// @Security BearerAuth
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} models.Invoice
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /guests/bookings/{id}/invoice [get]
func (h *BookingHandler) GetInvoice(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	bookingID := c.Param("id")
	invoice, err := h.Svc.GetInvoice(user.ID.String(), bookingID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, invoice)
}
