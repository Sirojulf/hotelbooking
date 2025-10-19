package handler

import (
	"hotelbooking/internal/models"
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RegisterGuestRequest struct {
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Email     string        `json:"email"`
	Password  string        `json:"password"`
	Phone     string        `json:"phone"`
	Gender    models.Gender `json:"gender"`
	Country   string        `json:"country"`
}

type RegisterGuestHandler struct {
	Svc service.RegisterGuestService
}

func NewRegisterGuestHandler(svc service.RegisterGuestService) *RegisterGuestHandler {
	return &RegisterGuestHandler{Svc: svc}
}

func (h *RegisterGuestHandler) Handle(c echo.Context) error {
	var req RegisterGuestRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	createdProfile, err := h.Svc.Execute(req.Email, req.Password, req.FirstName, req.LastName, req.Phone, req.Country, req.Gender)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, createdProfile)
}
