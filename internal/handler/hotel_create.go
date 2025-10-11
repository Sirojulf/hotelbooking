package handler

import (
	"errors"
	"hotelbooking/internal/models"
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supabase-community/gotrue-go/types"
)

type CreateHotelHandler struct {
	Svc service.CreateHotelService
}

func NewCreateHotelHandler(svc service.CreateHotelService) *CreateHotelHandler {
	return &CreateHotelHandler{Svc: svc}
}

func (h *CreateHotelHandler) Handle(c echo.Context) error {
	user, ok := c.Get("user").(*types.UserResponse)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid user data in token"})
	}

	userRole, err := getUserRoleFromMetadata(user.UserMetadata)
	if err != nil || userRole != "admin" {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Hanya admin yang dapat mengakses"})
	}

	var hotel models.Hotel
	if err := c.Bind(&hotel); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	createdHotel, err := h.Svc.Execute(hotel)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, createdHotel)
}

func getUserRoleFromMetadata(metadata map[string]interface{}) (string, error) {
	role, ok := metadata["user_role"].(string)
	if !ok {
		return "", errors.New("user_role tidak ditemukan di metadata")
	}
	return role, nil
}
