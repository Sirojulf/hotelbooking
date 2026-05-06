package handler

import (
	"hotelbooking/internal/middleware"
	"hotelbooking/internal/service"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateProfileRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	HotelID  string `json:"hotel_id"`
}

type UpdateProfileRequest struct {
	Role    string `json:"role"`
	HotelID string `json:"hotel_id"`
}

type AdminHandler struct {
	profileSvc     service.ProfileService
	reservationSvc service.ReservationService
}

func NewAdminHandler(profileSvc service.ProfileService, resSvc service.ReservationService) *AdminHandler {
	return &AdminHandler{profileSvc: profileSvc, reservationSvc: resSvc}
}

// @Summary Login admin/staff
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body AdminLoginRequest true "Login credentials"
// @Success 200 {object} handler.AdminLoginResponseDoc
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/admin/login [post]
func (h *AdminHandler) Login(c echo.Context) error {
	var req AdminLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	resp, err := h.profileSvc.Login(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

// @Summary Create staff profile
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateProfileRequest true "Create profile"
// @Success 201 {object} models.Profile
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin/users [post]
func (h *AdminHandler) CreateProfile(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req CreateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	if !middleware.IsSuperAdmin(profile) {
		if req.HotelID == "" {
			req.HotelID = profile.HotelID.String()
		} else if req.HotelID != profile.HotelID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "tidak dapat membuat profil untuk hotel lain"})
		}
	}
	created, err := h.profileSvc.CreateStaffProfile(service.CreateProfileInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     req.Role,
		HotelID:  req.HotelID,
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, created)
}

// @Summary List staff profiles
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param hotel_id query string false "Hotel ID"
// @Success 200 {array} models.Profile
// @Failure 401 {object} map[string]string
// @Router /admin/users [get]
func (h *AdminHandler) ListProfiles(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	hotelID := c.QueryParam("hotel_id")
	if !middleware.IsSuperAdmin(profile) {
		hotelID = profile.HotelID.String()
	}
	profiles, err := h.profileSvc.ListProfiles(hotelID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, profiles)
}

// @Summary Update staff profile
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param payload body UpdateProfileRequest true "Update profile"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /admin/users/{id} [put]
func (h *AdminHandler) UpdateProfile(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if !middleware.IsSuperAdmin(profile) {
		target, err := h.profileSvc.GetProfileByID(id)
		if err != nil || target.HotelID == nil || target.HotelID.String() != profile.HotelID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "akses ditolak"})
		}
	}
	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	if req.Role != "" {
		if err := h.profileSvc.UpdateRole(id, req.Role); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
		}
	}
	if req.HotelID != "" && middleware.IsSuperAdmin(profile) {
		if err := h.profileSvc.UpdateHotel(id, req.HotelID); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "updated"})
}

// @Summary List reservations (admin)
// @Tags Reservations
// @Security BearerAuth
// @Produce json
// @Param hotel_id query string false "Hotel ID"
// @Param status query string false "Payment status"
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} models.Reservation
// @Failure 401 {object} map[string]string
// @Router /admin/reservations [get]
func (h *AdminHandler) ListReservations(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	hotelID := c.QueryParam("hotel_id")
	if !middleware.IsSuperAdmin(profile) {
		hotelID = profile.HotelID.String()
	}
	status := c.QueryParam("status")
	start := c.QueryParam("start")
	end := c.QueryParam("end")

	var startTime, endTime time.Time
	var err error
	if start != "" {
		if startTime, err = time.Parse("2006-01-02", start); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "format start date tidak valid"})
		}
	}
	if end != "" {
		if endTime, err = time.Parse("2006-01-02", end); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "format end date tidak valid"})
		}
	}

	reservations, err := h.reservationSvc.ListReservations(hotelID, status, startTime, endTime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, reservations)
}

type UpdateReservationRequest struct {
	PaymentStatus string `json:"payment_status"`
	AINotes       string `json:"ai_notes"`
	CheckedInAt   string `json:"checked_in_at"`
	CheckedOutAt  string `json:"checked_out_at"`
}

// @Summary Update reservation (admin)
// @Tags Reservations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Reservation ID"
// @Param payload body UpdateReservationRequest true "Update reservation"
// @Success 200 {object} models.Reservation
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /admin/reservations/{id} [put]
func (h *AdminHandler) UpdateReservation(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if !middleware.IsSuperAdmin(profile) {
		res, err := h.reservationSvc.GetReservationByID(id)
		if err != nil || res.HotelID.String() != profile.HotelID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "akses ditolak"})
		}
	}
	var req UpdateReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	updates := map[string]any{}
	if req.PaymentStatus != "" {
		updates["payment_status"] = req.PaymentStatus
	}
	if req.AINotes != "" {
		updates["ai_notes"] = req.AINotes
	}
	if req.CheckedInAt != "" {
		updates["checked_in_at"] = req.CheckedInAt
	}
	if req.CheckedOutAt != "" {
		updates["checked_out_at"] = req.CheckedOutAt
	}
	res, err := h.reservationSvc.UpdateReservation(id, updates)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}
