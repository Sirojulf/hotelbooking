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

type CreateRoomRequest struct {
	PropertyID string `json:"property_id"`
	RoomTypeID string `json:"room_type_id"`
	RoomNumber string `json:"room_number"`
}

func (h *InventoryHandler) CreateRoom(c echo.Context) error {
	var req CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	result, err := h.Svc.CreateRoom(req.PropertyID, req.RoomTypeID, req.RoomNumber)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

type CreateRoomTypeRequest struct {
	PropertyID  string   `json:"property_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	BasePrice   float64  `json:"base_price"`
	Capacity    int      `json:"capacity"`
	Facilities  []string `json:"facilities"`
}

func (h *InventoryHandler) CreateRoomType(c echo.Context) error {
	var req CreateRoomTypeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	result, err := h.Svc.CreateRoomType(req.PropertyID, req.Name, req.Description, req.BasePrice, req.Capacity, req.Facilities)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}
