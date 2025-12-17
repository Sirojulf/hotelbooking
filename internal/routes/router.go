package routes

import (
	"hotelbooking/internal/handler"
	"hotelbooking/internal/middleware"
	"hotelbooking/internal/repository"
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	// Health check
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hotel Booking API is running!")
	})

	api := e.Group("/api/v1")

	// ======================
	// REPOSITORIES
	// ======================
	guestRepo := repository.NewGuestRepo()
	adminRepo := repository.NewAdminRepo()
	propertyRepo := repository.NewPropertyRepo()
	bookingRepo := repository.NewBookingRepo()

	// ======================
	// SERVICES (DOMAIN BASED)
	// ======================
	// Guest domain: auth + experience (search hotel, bookings, profile)
	guestSvc := service.NewGuestService(guestRepo, propertyRepo, bookingRepo)

	// Admin domain: login + (nanti) manajemen admin
	adminSvc := service.NewAdminService(adminRepo)

	// Inventory domain (admin kelola hotel/room/room-type)
	inventorySvc := service.NewInventoryService(propertyRepo)

	// ======================
	// HANDLERS
	// ======================
	guestHandler := handler.NewGuestHandler(guestSvc)
	adminHandler := handler.NewAdminHandler(adminSvc)
	inventoryHandler := handler.NewInventoryHandler(inventorySvc)

	// ======================
	// PUBLIC ROUTES
	// ======================

	// Auth Guest
	api.POST("/auth/guest/register", guestHandler.Register)
	api.POST("/auth/guest/login", guestHandler.Login)

	// Auth Admin
	api.POST("/auth/admin/login", adminHandler.Login)

	// Guest Experience (tanpa login: explore hotel)
	api.GET("/hotels", guestHandler.SearchHotels)       // ?city=Jakarta
	api.GET("/hotels/:id", guestHandler.GetHotelDetail) // detail 1 hotel

	// ======================
	// PROTECTED ROUTES (BUTUH TOKEN)
	// ======================

	// Group khusus Guest (butuh AuthMiddleware)
	guestGroup := api.Group("/guests")
	guestGroup.Use(middleware.AuthMiddleware)
	guestGroup.GET("/bookings", guestHandler.GetMyBookings)
	guestGroup.GET("/me", guestHandler.GetMyProfile)

	// Group khusus Admin (butuh AuthMiddleware)
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware)

	// Fitur admin untuk manage hotel & inventory
	adminGroup.POST("/hotels", inventoryHandler.CreateHotel)
	adminGroup.POST("/room-types", inventoryHandler.CreateRoomType)
	adminGroup.POST("/rooms", inventoryHandler.CreateRoom)
}
