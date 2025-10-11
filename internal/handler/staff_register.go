package handler

import (
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RegisterStaffRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type RegisterStaffHandler struct {
	Svc service.RegisterStaffService
}

func NewRegisterStaffHandler(svc service.RegisterStaffService) *RegisterStaffHandler {
	return &RegisterStaffHandler{Svc: svc}
}

func (h *RegisterStaffHandler) Handle(c echo.Context) error {
	var req RegisterStaffRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	createdProfile, err := h.Svc.Execute(req.Email, req.Password, req.FullName, req.Role)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, createdProfile)
}
