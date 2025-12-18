// internal/handler/admin_handler.go
package handler

import (
	"hotelbooking/internal/models"
	"hotelbooking/internal/service"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"hotelbooking/internal/middleware"
)

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAdminRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	PropertyID string `json:"property_id"` // bisa kosong kalau mau super admin
	Role       string `json:"role"`
}

type AdminHandler struct {
	Svc        service.AdminService
	BookingSvc service.BookingService
}

func NewAdminHandler(svc service.AdminService, bookingSvc service.BookingService) *AdminHandler {
	return &AdminHandler{Svc: svc, BookingSvc: bookingSvc}
}

// @Summary Login admin
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body AdminLoginRequest true "Admin login"
// @Success 200 {object} AdminLoginResponseDoc
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/admin/login [post]
func (h *AdminHandler) Login(c echo.Context) error {
	var req AdminLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	resp, err := h.Svc.Login(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// @Summary Create admin user
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateAdminRequest true "Create admin"
// @Success 201 {object} models.Admin
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/users [post]
func (h *AdminHandler) CreateAdmin(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req CreateAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}
	if admin.PropertyID != nil {
		if req.PropertyID == "" {
			req.PropertyID = admin.PropertyID.String()
		} else if req.PropertyID != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	admin, err := h.Svc.CreateAdmin(service.CreateAdminInput{
		Email:      req.Email,
		Password:   req.Password,
		PropertyID: req.PropertyID,
		Role:       req.Role,
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, admin)
}

// @Summary Activate admin user
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "Admin ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/users/{id}/activate [post]
func (h *AdminHandler) Activate(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if admin.PropertyID != nil {
		target, err := h.Svc.GetAdminByID(c.Param("id"))
		if err != nil || target.PropertyID == nil || target.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if err := h.Svc.ActivateAdmin(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "activated"})
}

// @Summary Deactivate admin user
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "Admin ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/users/{id}/deactivate [post]
func (h *AdminHandler) Deactivate(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if admin.PropertyID != nil {
		target, err := h.Svc.GetAdminByID(c.Param("id"))
		if err != nil || target.PropertyID == nil || target.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if err := h.Svc.DeactivateAdmin(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "deactivated"})
}

type UpdateAdminRequest struct {
	Role       string `json:"role"`
	PropertyID string `json:"property_id"`
	IsActive   *bool  `json:"is_active"`
}

// @Summary Update admin user
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Admin ID"
// @Param payload body UpdateAdminRequest true "Update admin"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/users/{id} [put]
func (h *AdminHandler) UpdateAdmin(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if admin.PropertyID != nil {
		target, err := h.Svc.GetAdminByID(id)
		if err != nil || target.PropertyID == nil || target.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	var req UpdateAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}
	if admin.PropertyID != nil && req.PropertyID != "" && req.PropertyID != admin.PropertyID.String() {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
	}
	if req.Role != "" {
		if err := h.Svc.UpdateRole(id, req.Role); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
		}
	}
	if req.PropertyID != "" {
		if err := h.Svc.UpdateProperty(id, req.PropertyID); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
		}
	}
	if req.IsActive != nil {
		if *req.IsActive {
			if err := h.Svc.ActivateAdmin(id); err != nil {
				return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
			}
		} else {
			if err := h.Svc.DeactivateAdmin(id); err != nil {
				return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
			}
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "updated"})
}

// @Summary List admin users
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param property_id query string false "Property ID"
// @Success 200 {array} models.Admin
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users [get]
func (h *AdminHandler) ListAdmins(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	propertyID := c.QueryParam("property_id")
	if admin.PropertyID != nil {
		propertyID = admin.PropertyID.String()
	}
	admins, err := h.Svc.ListAdmins(propertyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, admins)
}

// -------- Booking oversight (domain admin) --------

// @Summary List bookings
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Param property_id query string false "Property ID"
// @Param status query string false "Booking status"
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} models.Booking
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/bookings [get]
func (h *AdminHandler) ListBookings(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	propertyID := c.QueryParam("property_id")
	if admin.PropertyID != nil {
		propertyID = admin.PropertyID.String()
	}
	status := c.QueryParam("status")
	start := c.QueryParam("start")
	end := c.QueryParam("end")

	var startTime, endTime time.Time
	var err error
	if start != "" {
		startTime, err = time.Parse("2006-01-02", start)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid start date"})
		}
	}
	if end != "" {
		endTime, err = time.Parse("2006-01-02", end)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid end date"})
		}
	}

	bookings, err := h.BookingSvc.ListBookings(propertyID, status, startTime, endTime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, bookings)
}

type UpdateBookingStatusRequest struct {
	Status       models.BookingStatus `json:"status"`
	Note         string               `json:"note"`
	RefundAmount float64              `json:"refund_amount"`
}

// @Summary Update booking status
// @Tags Bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param payload body UpdateBookingStatusRequest true "Update booking status"
// @Success 200 {object} models.Booking
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/bookings/{id}/status [put]
func (h *AdminHandler) UpdateBookingStatus(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if admin.PropertyID != nil {
		booking, err := h.BookingSvc.GetBookingByID(id)
		if err != nil || booking.PropertyID == nil || booking.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	var req UpdateBookingStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}
	booking, err := h.BookingSvc.UpdateStatus(id, req.Status, req.Note, req.RefundAmount)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, booking)
}
