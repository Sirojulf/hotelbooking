// internal/handler/admin_handler.go
package handler

import (
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAdminRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	PropertyID string `json:"property_id"` // bisa kosong kalau mau super admin
}

type AdminHandler struct {
	Svc service.AdminService
}

func NewAdminHandler(svc service.AdminService) *AdminHandler {
	return &AdminHandler{Svc: svc}
}

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
