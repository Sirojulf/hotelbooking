// internal/handler/guest_handler.go
package handler

import (
	"hotelbooking/internal/models"
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supabase-community/gotrue-go/types"
)

// ===== Request DTO =====

type RegisterGuestRequest struct {
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Email     string        `json:"email"`
	Password  string        `json:"password"`
	Phone     string        `json:"phone"`
	Gender    models.Gender `json:"gender"`
	Country   string        `json:"country"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// ===== Handler =====

type GuestHandler struct {
	Svc service.GuestService
}

func NewGuestHandler(svc service.GuestService) *GuestHandler {
	return &GuestHandler{Svc: svc}
}

// POST /api/v1/auth/guest/register
func (h *GuestHandler) Register(c echo.Context) error {
	var req RegisterGuestRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	input := service.RegisterGuestInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
		Phone:     req.Phone,
		Gender:    req.Gender,
		Country:   req.Country,
	}

	guest, err := h.Svc.RegisterGuest(input)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, guest)
}

// POST /api/v1/auth/guest/login
func (h *GuestHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	session, err := h.Svc.LoginGuest(req.Login, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, session)
}

// GET /api/v1/hotels?city=Jakarta
func (h *GuestHandler) SearchHotels(c echo.Context) error {
	city := c.QueryParam("city")
	if city == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Parameter 'city' wajib diisi"})
	}

	result, err := h.Svc.SearchHotels(city)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// GET /api/v1/hotels/:id
func (h *GuestHandler) GetHotelDetail(c echo.Context) error {
	id := c.Param("id")
	result, err := h.Svc.GetHotelDetails(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Hotel tidak ditemukan"})
	}
	return c.JSON(http.StatusOK, result)
}

// GET /api/v1/guests/bookings  (perlu Auth middleware)
func (h *GuestHandler) GetMyBookings(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	bookings, err := h.Svc.GetMyBookings(user.ID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, bookings)
}

// GET /api/v1/guests/me
func (h *GuestHandler) GetMyProfile(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	profile, err := h.Svc.GetMyProfile(user.ID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, profile)
}
