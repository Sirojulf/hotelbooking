package handler

import (
	"hotelbooking/internal/middleware"
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type InventoryHandler struct {
	svc service.InventoryService
}

func NewInventoryHandler(svc service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

// ─── Hotels ──────────────────────────────────────────────────────────────────

// @Summary Create hotel
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body service.CreateHotelInput true "Create hotel"
// @Success 201 {object} models.Hotel
// @Router /admin/hotels [post]
func (h *InventoryHandler) CreateHotel(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if !middleware.IsSuperAdmin(profile) {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "hanya super admin yang dapat membuat hotel"})
	}
	var req service.CreateHotelInput
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	result, err := h.svc.CreateHotel(req)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, result)
}

// @Summary Update hotel
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Hotel ID"
// @Param payload body service.UpdateHotelInput true "Update hotel"
// @Success 200 {object} models.Hotel
// @Router /admin/hotels/{id} [put]
func (h *InventoryHandler) UpdateHotel(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if !middleware.IsSuperAdmin(profile) && profile.HotelID.String() != id {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "akses ditolak"})
	}
	var req service.UpdateHotelInput
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	res, err := h.svc.UpdateHotel(id, req)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

// @Summary Delete hotel
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Hotel ID"
// @Success 204
// @Router /admin/hotels/{id} [delete]
func (h *InventoryHandler) DeleteHotel(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if !middleware.IsSuperAdmin(profile) {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "hanya super admin yang dapat menghapus hotel"})
	}
	if err := h.svc.DeleteHotel(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// @Summary List hotels
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param q query string false "Search query"
// @Success 200 {array} models.Hotel
// @Router /admin/hotels [get]
func (h *InventoryHandler) ListHotels(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if !middleware.IsSuperAdmin(profile) {
		hotel, err := h.svc.GetHotelByID(profile.HotelID.String())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, []any{hotel})
	}
	hotels, err := h.svc.ListHotels(c.QueryParam("q"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, hotels)
}

// ─── Room Types ───────────────────────────────────────────────────────────────

// @Summary Create room type
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body service.CreateRoomTypeInput true "Create room type"
// @Success 201 {object} models.RoomType
// @Router /admin/room-types [post]
func (h *InventoryHandler) CreateRoomType(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req service.CreateRoomTypeInput
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	if !middleware.IsSuperAdmin(profile) {
		if req.HotelID == "" {
			req.HotelID = profile.HotelID.String()
		} else if req.HotelID != profile.HotelID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "akses ditolak"})
		}
	}
	result, err := h.svc.CreateRoomType(req)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, result)
}

// @Summary Update room type
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Room Type ID"
// @Param payload body service.UpdateRoomTypeInput true "Update room type"
// @Success 200 {object} models.RoomType
// @Router /admin/room-types/{id} [put]
func (h *InventoryHandler) UpdateRoomType(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if !middleware.IsSuperAdmin(profile) {
		rt, err := h.svc.GetRoomTypeByID(id)
		if err != nil || rt.HotelID.String() != profile.HotelID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "akses ditolak"})
		}
	}
	var req service.UpdateRoomTypeInput
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	res, err := h.svc.UpdateRoomType(id, req)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

// @Summary Delete room type
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Room Type ID"
// @Success 204
// @Router /admin/room-types/{id} [delete]
func (h *InventoryHandler) DeleteRoomType(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if !middleware.IsSuperAdmin(profile) {
		rt, err := h.svc.GetRoomTypeByID(c.Param("id"))
		if err != nil || rt.HotelID.String() != profile.HotelID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "akses ditolak"})
		}
	}
	if err := h.svc.DeleteRoomType(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// @Summary List room types
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param hotel_id query string false "Hotel ID"
// @Success 200 {array} models.RoomType
// @Router /admin/room-types [get]
func (h *InventoryHandler) ListRoomTypes(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	hotelID := c.QueryParam("hotel_id")
	if !middleware.IsSuperAdmin(profile) {
		hotelID = profile.HotelID.String()
	}
	res, err := h.svc.ListRoomTypes(hotelID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

// ─── Rooms ────────────────────────────────────────────────────────────────────

// @Summary Create room
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body service.CreateRoomInput true "Create room"
// @Success 201 {object} models.Room
// @Router /admin/rooms [post]
func (h *InventoryHandler) CreateRoom(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req service.CreateRoomInput
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	if !middleware.IsSuperAdmin(profile) {
		if req.HotelID == "" {
			req.HotelID = profile.HotelID.String()
		} else if req.HotelID != profile.HotelID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "akses ditolak"})
		}
	}
	result, err := h.svc.CreateRoom(req)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, result)
}

// @Summary Update room
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Room ID"
// @Param payload body service.UpdateRoomInput true "Update room"
// @Success 200 {object} models.Room
// @Router /admin/rooms/{id} [put]
func (h *InventoryHandler) UpdateRoom(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	id := c.Param("id")
	if !middleware.IsSuperAdmin(profile) {
		room, err := h.svc.GetRoomByID(id)
		if err != nil || room.HotelID.String() != profile.HotelID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "akses ditolak"})
		}
	}
	var req service.UpdateRoomInput
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "request tidak valid"})
	}
	res, err := h.svc.UpdateRoom(id, req)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

// @Summary Delete room
// @Tags Inventory
// @Security BearerAuth
// @Param id path string true "Room ID"
// @Success 204
// @Router /admin/rooms/{id} [delete]
func (h *InventoryHandler) DeleteRoom(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	if !middleware.IsSuperAdmin(profile) {
		room, err := h.svc.GetRoomByID(c.Param("id"))
		if err != nil || room.HotelID.String() != profile.HotelID.String() {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "akses ditolak"})
		}
	}
	if err := h.svc.DeleteRoom(c.Param("id")); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// @Summary List rooms
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param hotel_id query string false "Hotel ID"
// @Param room_type_id query string false "Room Type ID"
// @Success 200 {array} models.Room
// @Router /admin/rooms [get]
func (h *InventoryHandler) ListRooms(c echo.Context) error {
	profile, ok := middleware.GetProfileFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	hotelID := c.QueryParam("hotel_id")
	if !middleware.IsSuperAdmin(profile) {
		hotelID = profile.HotelID.String()
	}
	res, err := h.svc.ListRooms(hotelID, c.QueryParam("room_type_id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}
