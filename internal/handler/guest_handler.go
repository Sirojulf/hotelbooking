package handler

import (
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supabase-community/gotrue-go/types"
)

type RegisterGuestRequest struct {
	HotelID     string `json:"hotel_id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
	Title       string `json:"title"`
}

type GuestLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type GuestHandler struct {
	svc service.GuestService
}

func NewGuestHandler(svc service.GuestService) *GuestHandler {
	return &GuestHandler{svc: svc}
}

// @Summary Register guest
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body RegisterGuestRequest true "Register guest"
// @Success 201 {object} models.Guest
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /auth/guest/register [post]
func (h *GuestHandler) Register(c echo.Context) error {
	var req RegisterGuestRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	guest, err := h.svc.RegisterGuest(service.RegisterGuestInput{
		HotelID:     req.HotelID,
		FullName:    req.FullName,
		Email:       req.Email,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		Title:       req.Title,
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, guest)
}

// @Summary Login guest
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body GuestLoginRequest true "Login"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/guest/login [post]
func (h *GuestHandler) Login(c echo.Context) error {
	var req GuestLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	session, err := h.svc.LoginGuest(req.Login, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, session)
}

// @Summary Search hotels
// @Tags Hotels
// @Produce json
// @Param q query string false "Search query (name/address)"
// @Success 200 {array} models.Hotel
// @Failure 500 {object} map[string]string
// @Router /hotels [get]
func (h *GuestHandler) SearchHotels(c echo.Context) error {
	q := c.QueryParam("q")
	result, err := h.svc.SearchHotels(q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

// @Summary Get hotel detail
// @Tags Hotels
// @Produce json
// @Param id path string true "Hotel ID"
// @Success 200 {object} models.HotelDetailResponse
// @Failure 404 {object} map[string]string
// @Router /hotels/{id} [get]
func (h *GuestHandler) GetHotelDetail(c echo.Context) error {
	result, err := h.svc.GetHotelDetails(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "hotel tidak ditemukan"})
	}
	return c.JSON(http.StatusOK, result)
}

// @Summary Get my reservations
// @Tags Guests
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Reservation
// @Failure 401 {object} map[string]string
// @Router /guests/reservations [get]
func (h *GuestHandler) GetMyReservations(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	reservations, err := h.svc.GetMyReservations(user.ID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, reservations)
}

// @Summary Get my profile
// @Tags Guests
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.Guest
// @Failure 401 {object} map[string]string
// @Router /guests/me [get]
func (h *GuestHandler) GetMyProfile(c echo.Context) error {
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	profile, err := h.svc.GetMyProfile(user.ID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, profile)
}
