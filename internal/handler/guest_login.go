package handler

import (
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginGuestHandler struct {
	Svc service.LoginGuestService
}

func NewLoginGuestHandler(svc service.LoginGuestService) *LoginGuestHandler {
	return &LoginGuestHandler{Svc: svc}
}

func (h *LoginGuestHandler) Handle(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	session, err := h.Svc.Execute(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, session)
}
