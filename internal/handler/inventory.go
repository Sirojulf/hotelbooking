package handler

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/models"
	"hotelbooking/internal/middleware"
	"hotelbooking/internal/service"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CreateHotelRequest struct {
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	City      string   `json:"city"`
	HotelCode string   `json:"hotel_code"`
	Facilities []string `json:"facilities"`
}

type UpdateHotelRequest struct {
	Name               string   `json:"name"`
	Address            string   `json:"address"`
	City               string   `json:"city"`
	Facilities         []string `json:"facilities"`
	CheckInTime        string   `json:"checkin_time"`
	CheckOutTime       string   `json:"checkout_time"`
	CancellationPolicy string   `json:"cancellation_policy"`
}

type InventoryHandler struct {
	Svc service.InventoryService
}

func NewInventoryHandler(svc service.InventoryService) *InventoryHandler {
	return &InventoryHandler{Svc: svc}
}

// @Summary Create hotel
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateHotelRequest true "Create hotel"
// @Success 201 {object} models.Properties
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/hotels [post]
func (h *InventoryHandler) CreateHotel(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if admin.PropertyID != nil {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
	}
	var req CreateHotelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	result, err := h.Svc.CreateHotel(req.Name, req.Address, req.City, req.HotelCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

// @Summary Update hotel
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Hotel ID"
// @Param payload body UpdateHotelRequest true "Update hotel"
// @Success 200 {object} models.Properties
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/hotels/{id} [put]
func (h *InventoryHandler) UpdateHotel(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if admin.PropertyID != nil && admin.PropertyID.String() != id {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
	}
	var req UpdateHotelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	res, err := h.Svc.UpdateHotel(id, req.Name, req.Address, req.City, req.Facilities, req.CheckInTime, req.CheckOutTime, req.CancellationPolicy)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

// @Summary Delete hotel
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Hotel ID"
// @Success 204 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/hotels/{id} [delete]
func (h *InventoryHandler) DeleteHotel(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if admin.PropertyID != nil && admin.PropertyID.String() != c.Param("id") {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
	}
	if err := h.Svc.DeleteHotel(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// @Summary List hotels
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param city query string false "City filter"
// @Success 200 {array} models.Properties
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/hotels [get]
func (h *InventoryHandler) ListHotels(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if admin.PropertyID != nil {
		property, err := h.Svc.GetHotelByID(admin.PropertyID.String())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, []models.Properties{*property})
	}
	res, err := h.Svc.ListHotels(c.QueryParam("city"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

type CreateRoomTypeRequest struct {
	PropertyID  string   `json:"property_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	BasePrice   float64  `json:"base_price"`
	Capacity    int      `json:"capacity"`
	Facilities  []string `json:"facilities"`
}

type UpdateRoomTypeRequest struct {
	PropertyID  string   `json:"property_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	BasePrice   float64  `json:"base_price"`
	Capacity    int      `json:"capacity"`
	Facilities  []string `json:"facilities"`
}

// @Summary Create room type
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateRoomTypeRequest true "Create room type"
// @Success 201 {object} models.RoomType
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/room-types [post]
func (h *InventoryHandler) CreateRoomType(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req CreateRoomTypeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	if admin.PropertyID != nil {
		if req.PropertyID == "" {
			req.PropertyID = admin.PropertyID.String()
		} else if req.PropertyID != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}

	result, err := h.Svc.CreateRoomType(req.PropertyID, req.Name, req.Description, req.BasePrice, req.Capacity, req.Facilities)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

// @Summary Update room type
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Room type ID"
// @Param payload body UpdateRoomTypeRequest true "Update room type"
// @Success 200 {object} models.RoomType
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/room-types/{id} [put]
func (h *InventoryHandler) UpdateRoomType(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if admin.PropertyID != nil {
		roomType, err := h.Svc.GetRoomTypeByID(id)
		if err != nil || roomType.PropertyID == nil || roomType.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	var req UpdateRoomTypeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	if admin.PropertyID != nil && req.PropertyID != "" && req.PropertyID != admin.PropertyID.String() {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
	}
	res, err := h.Svc.UpdateRoomType(id, req.PropertyID, req.Name, req.Description, req.BasePrice, req.Capacity, req.Facilities)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

// @Summary Delete room type
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Room type ID"
// @Success 204 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/room-types/{id} [delete]
func (h *InventoryHandler) DeleteRoomType(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if admin.PropertyID != nil {
		roomType, err := h.Svc.GetRoomTypeByID(c.Param("id"))
		if err != nil || roomType.PropertyID == nil || roomType.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if err := h.Svc.DeleteRoomType(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// @Summary List room types
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param property_id query string false "Property ID"
// @Success 200 {array} models.RoomType
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/room-types [get]
func (h *InventoryHandler) ListRoomTypes(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	propertyID := c.QueryParam("property_id")
	if admin.PropertyID != nil {
		propertyID = admin.PropertyID.String()
	}
	res, err := h.Svc.ListRoomTypes(propertyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

type CreateRoomRequest struct {
	PropertyID string `json:"property_id"`
	RoomTypeID string `json:"room_type_id"`
	RoomNumber string `json:"room_number"`
}

type UpdateRoomRequest struct {
	PropertyID         string                   `json:"property_id"`
	RoomTypeID         string                   `json:"room_type_id"`
	RoomNumber         string                   `json:"room_number"`
	Status             models.RoomStatus        `json:"status"`
	HousekeepingStatus models.HousekeepingStatus `json:"housekeeping_status"`
}

// @Summary Create room
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateRoomRequest true "Create room"
// @Success 201 {object} models.Room
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/rooms [post]
func (h *InventoryHandler) CreateRoom(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	if admin.PropertyID != nil {
		if req.PropertyID == "" {
			req.PropertyID = admin.PropertyID.String()
		} else if req.PropertyID != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if req.RoomTypeID != "" {
		roomType, err := h.Svc.GetRoomTypeByID(req.RoomTypeID)
		if err != nil || roomType.PropertyID == nil || (admin.PropertyID != nil && roomType.PropertyID.String() != admin.PropertyID.String()) {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	result, err := h.Svc.CreateRoom(req.PropertyID, req.RoomTypeID, req.RoomNumber)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

// @Summary Update room
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Room ID"
// @Param payload body UpdateRoomRequest true "Update room"
// @Success 200 {object} models.Room
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/rooms/{id} [put]
func (h *InventoryHandler) UpdateRoom(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if admin.PropertyID != nil {
		room, err := h.Svc.GetRoomByID(id)
		if err != nil || room.PropertyID == nil || room.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	var req UpdateRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	if admin.PropertyID != nil && req.PropertyID != "" && req.PropertyID != admin.PropertyID.String() {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
	}
	if req.RoomTypeID != "" {
		roomType, err := h.Svc.GetRoomTypeByID(req.RoomTypeID)
		if err != nil || roomType.PropertyID == nil || (admin.PropertyID != nil && roomType.PropertyID.String() != admin.PropertyID.String()) {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	res, err := h.Svc.UpdateRoom(id, req.PropertyID, req.RoomTypeID, req.RoomNumber, req.Status, req.HousekeepingStatus)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

// @Summary Delete room
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Room ID"
// @Success 204 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/rooms/{id} [delete]
func (h *InventoryHandler) DeleteRoom(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if admin.PropertyID != nil {
		room, err := h.Svc.GetRoomByID(c.Param("id"))
		if err != nil || room.PropertyID == nil || room.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if err := h.Svc.DeleteRoom(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// @Summary List rooms
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param property_id query string false "Property ID"
// @Param room_type_id query string false "Room type ID"
// @Success 200 {array} models.Room
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/rooms [get]
func (h *InventoryHandler) ListRooms(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	propertyID := c.QueryParam("property_id")
	if admin.PropertyID != nil {
		propertyID = admin.PropertyID.String()
	}
	res, err := h.Svc.ListRooms(propertyID, c.QueryParam("room_type_id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

type RoomRateRequest struct {
	RoomID           string          `json:"room_id"`
	Dates            []string        `json:"dates"`
	AvailableRooms   int             `json:"available_rooms"`
	LinearRate       *float64        `json:"linear_rate"`
	NonLinearRate    json.RawMessage `json:"non_linear_rate"`
	MinNights        int             `json:"min_nights"`
	MaxNights        int             `json:"max_nights"`
	StopSell         bool            `json:"stop_sell"`
	CloseOnArrival   bool            `json:"close_on_arrival"`
	CloseOnDeparture bool            `json:"close_on_departure"`
}

// @Summary Set room rates
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param room_id path string true "Room ID"
// @Param payload body RoomRateRequestDoc true "Room rate payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/rooms/{room_id}/rates [post]
func (h *InventoryHandler) SetRoomRates(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req RoomRateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	if req.RoomID == "" || len(req.Dates) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "room_id and dates are required"})
	}
	roomUUID, err := uuid.Parse(req.RoomID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid room_id"})
	}
	if admin.PropertyID != nil {
		room, err := h.Svc.GetRoomByID(req.RoomID)
		if err != nil || room.PropertyID == nil || room.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	var rates []models.RoomRate
	for _, d := range req.Dates {
		parsed, parseErr := time.Parse("2006-01-02", d)
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": fmt.Sprintf("invalid date: %s", d)})
		}
		rates = append(rates, models.RoomRate{
			ID:               uuid.New(),
			RoomID:           &roomUUID,
			Date:             parsed,
			AvailableRooms:   req.AvailableRooms,
			LinearRate:       req.LinearRate,
			NonLinearRate:    req.NonLinearRate,
			MinNights:        req.MinNights,
			MaxNights:        req.MaxNights,
			StopSell:         req.StopSell,
			CloseOnArrival:   req.CloseOnArrival,
			CloseOnDeparture: req.CloseOnDeparture,
			CreatedAt:        time.Now(),
		})
	}
	if err := h.Svc.SetRoomRates(rates); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
}

// @Summary Get room rates
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param room_id path string true "Room ID"
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} RoomRateDoc
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/rooms/{room_id}/rates [get]
func (h *InventoryHandler) GetRoomRates(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	roomID := c.Param("room_id")
	if admin.PropertyID != nil {
		room, err := h.Svc.GetRoomByID(roomID)
		if err != nil || room.PropertyID == nil || room.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	start := c.QueryParam("start")
	end := c.QueryParam("end")
	rates, err := h.Svc.GetRoomRates(roomID, start, end)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, rates)
}

type PhotoRequest struct {
	URL     string `json:"url"`
	Caption string `json:"caption"`
}

// @Summary Add property photo
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param property_id path string true "Property ID"
// @Param payload body PhotoRequest true "Photo payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/hotels/{property_id}/photos [post]
func (h *InventoryHandler) AddPropertyPhoto(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	propertyID := c.Param("property_id")
	if admin.PropertyID != nil && admin.PropertyID.String() != propertyID {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
	}
	var req PhotoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	if err := h.Svc.AddPropertyPhoto(propertyID, req.URL, req.Caption); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, echo.Map{"status": "ok"})
}

// @Summary List property photos
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param property_id path string true "Property ID"
// @Success 200 {array} models.PropertyPhoto
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/hotels/{property_id}/photos [get]
func (h *InventoryHandler) ListPropertyPhotos(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	propertyID := c.Param("property_id")
	if admin.PropertyID != nil && admin.PropertyID.String() != propertyID {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
	}
	res, err := h.Svc.ListPropertyPhotos(propertyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

// @Summary Delete property photo
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Photo ID"
// @Success 204 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/photos/property/{id} [delete]
func (h *InventoryHandler) DeletePropertyPhoto(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if admin.PropertyID != nil {
		photo, err := h.Svc.GetPropertyPhotoByID(c.Param("id"))
		if err != nil || photo.PropertyID == nil || photo.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if err := h.Svc.DeletePropertyPhoto(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// @Summary Add room photo
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param property_id query string false "Property ID"
// @Param room_type_id query string false "Room type ID"
// @Param room_id query string false "Room ID"
// @Param payload body PhotoRequest true "Photo payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/room-photos [post]
func (h *InventoryHandler) AddRoomPhoto(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req PhotoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	propertyID := c.QueryParam("property_id")
	roomTypeID := c.QueryParam("room_type_id")
	roomID := c.QueryParam("room_id")
	if admin.PropertyID != nil {
		if propertyID == "" {
			propertyID = admin.PropertyID.String()
		} else if propertyID != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if roomTypeID != "" {
		roomType, err := h.Svc.GetRoomTypeByID(roomTypeID)
		if err != nil || roomType.PropertyID == nil || (admin.PropertyID != nil && roomType.PropertyID.String() != admin.PropertyID.String()) {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if roomID != "" {
		room, err := h.Svc.GetRoomByID(roomID)
		if err != nil || room.PropertyID == nil || (admin.PropertyID != nil && room.PropertyID.String() != admin.PropertyID.String()) {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if err := h.Svc.AddRoomPhoto(propertyID, roomTypeID, roomID, req.URL, req.Caption); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, echo.Map{"status": "ok"})
}

// @Summary List room photos
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param room_type_id query string false "Room type ID"
// @Param room_id query string false "Room ID"
// @Success 200 {array} models.RoomPhoto
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/room-photos [get]
func (h *InventoryHandler) ListRoomPhotos(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	roomTypeID := c.QueryParam("room_type_id")
	roomID := c.QueryParam("room_id")
	if admin.PropertyID != nil {
		if roomTypeID == "" && roomID == "" {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "room_type_id or room_id is required"})
		}
		if roomTypeID != "" {
			roomType, err := h.Svc.GetRoomTypeByID(roomTypeID)
			if err != nil || roomType.PropertyID == nil || roomType.PropertyID.String() != admin.PropertyID.String() {
				return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
			}
		}
		if roomID != "" {
			room, err := h.Svc.GetRoomByID(roomID)
			if err != nil || room.PropertyID == nil || room.PropertyID.String() != admin.PropertyID.String() {
				return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
			}
		}
	}
	res, err := h.Svc.ListRoomPhotos(roomTypeID, roomID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

// @Summary Delete room photo
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Photo ID"
// @Success 204 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /admin/photos/room/{id} [delete]
func (h *InventoryHandler) DeleteRoomPhoto(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if admin.PropertyID != nil {
		photo, err := h.Svc.GetRoomPhotoByID(c.Param("id"))
		if err != nil || photo.PropertyID == nil || photo.PropertyID.String() != admin.PropertyID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden property access"})
		}
	}
	if err := h.Svc.DeleteRoomPhoto(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
