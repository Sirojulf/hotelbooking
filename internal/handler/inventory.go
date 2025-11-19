package handler

import (
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CreateHotelRequest struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	City      string `json:"city"`
	Country   string `json:"country"`
	HotelCode string `json:"hotel_code"`
}

type InventoryHandler struct {
	Svc service.InventoryService
}

func NewInventoryHandler(svc service.InventoryService) *InventoryHandler {
	return &InventoryHandler{Svc: svc}
}

func (h *InventoryHandler) CreateHotel(c echo.Context) error {
	var req CreateHotelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	result, err := h.Svc.CreateHotel(req.Name, req.Address, req.City, req.Country, req.HotelCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}
