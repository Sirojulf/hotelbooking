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

type AdminLoginHandler struct {
	Svc service.AdminLoginService
}

func NewAdminLoginHandler(svc service.AdminLoginService) *AdminLoginHandler {
	return &AdminLoginHandler{Svc: svc}
}

func (h *AdminLoginHandler) Login(c echo.Context) error {
	var req AdminLoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	session, err := h.Svc.Execute(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})

	}

	return c.JSON(http.StatusOK, session)
}
