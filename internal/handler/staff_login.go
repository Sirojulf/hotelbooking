package handler

import (
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type LoginStaffRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginStaffHandler struct {
	Svc service.LoginStaffService
}

func NewLoginStaffHandler(svc service.LoginStaffService) *LoginStaffHandler {
	return &LoginStaffHandler{Svc: svc}
}

func (h *LoginStaffHandler) Handle(c echo.Context) error {
	var req LoginStaffRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request bosy"})

	}

	session, err := h.Svc.Execute(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})

	}
	return c.JSON(http.StatusOK, session)
}
