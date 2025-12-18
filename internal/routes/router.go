package routes

import (
	"hotelbooking/internal/handler"
	"hotelbooking/internal/middleware"
	"hotelbooking/internal/repository"
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func SetupRoutes(e *echo.Echo) {
	// Health check
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hotel Booking API is running!")
	})
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api/v1")

	// ======================
	// REPOSITORIES
	// ======================
	guestRepo := repository.NewGuestRepo()
	adminRepo := repository.NewAdminRepo()
	propertyRepo := repository.NewPropertyRepo()
	bookingRepo := repository.NewBookingRepo()
	paymentRepo := repository.NewPaymentRepo()

	// ======================
	// SERVICES (DOMAIN BASED)
	// ======================
	// Guest domain: auth + experience (search hotel, bookings, profile)
	guestSvc := service.NewGuestService(guestRepo, propertyRepo, bookingRepo)

	// Admin domain: login + (nanti) manajemen admin
	adminSvc := service.NewAdminService(adminRepo)

	// Inventory domain (admin kelola hotel/room/room-type)
	inventorySvc := service.NewInventoryService(propertyRepo)
	bookingSvc := service.NewBookingService(bookingRepo, propertyRepo, paymentRepo)
	reportSvc := service.NewReportService(bookingRepo, propertyRepo)

	// ======================
	// HANDLERS
	// ======================
	guestHandler := handler.NewGuestHandler(guestSvc)
	adminHandler := handler.NewAdminHandler(adminSvc, bookingSvc)
	inventoryHandler := handler.NewInventoryHandler(inventorySvc)
	reportHandler := handler.NewReportHandler(reportSvc)
	bookingHandler := handler.NewBookingHandler(bookingSvc)

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
	api.GET("/rooms/:room_id/availability", bookingHandler.CheckAvailability)

	// ======================
	// PROTECTED ROUTES (BUTUH TOKEN)
	// ======================

	// Group khusus Guest (butuh AuthMiddleware)
	guestGroup := api.Group("/guests")
	guestGroup.Use(middleware.AuthMiddleware)
	guestGroup.GET("/bookings", guestHandler.GetMyBookings)
	guestGroup.GET("/me", guestHandler.GetMyProfile)
	guestGroup.POST("/bookings", bookingHandler.CreateBooking)
	guestGroup.POST("/bookings/:id/pay", bookingHandler.PayBooking)
	guestGroup.POST("/bookings/:id/cancel", bookingHandler.CancelBooking)
	guestGroup.GET("/bookings/:id/invoice", bookingHandler.GetInvoice)

	// Group khusus Admin (butuh AuthMiddleware)
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware)
	adminGroup.Use(middleware.AdminOnly(adminRepo))

	// Fitur admin untuk manage hotel & inventory
	adminGroup.POST("/hotels", inventoryHandler.CreateHotel)
	adminGroup.GET("/hotels", inventoryHandler.ListHotels)
	adminGroup.PUT("/hotels/:id", inventoryHandler.UpdateHotel)
	adminGroup.DELETE("/hotels/:id", inventoryHandler.DeleteHotel)

	adminGroup.POST("/room-types", inventoryHandler.CreateRoomType)
	adminGroup.PUT("/room-types/:id", inventoryHandler.UpdateRoomType)
	adminGroup.DELETE("/room-types/:id", inventoryHandler.DeleteRoomType)
	adminGroup.GET("/room-types", inventoryHandler.ListRoomTypes)

	adminGroup.POST("/rooms", inventoryHandler.CreateRoom)
	adminGroup.PUT("/rooms/:id", inventoryHandler.UpdateRoom)
	adminGroup.DELETE("/rooms/:id", inventoryHandler.DeleteRoom)
	adminGroup.GET("/rooms", inventoryHandler.ListRooms)

	adminGroup.POST("/rooms/:room_id/rates", inventoryHandler.SetRoomRates)
	adminGroup.GET("/rooms/:room_id/rates", inventoryHandler.GetRoomRates)

	adminGroup.POST("/hotels/:property_id/photos", inventoryHandler.AddPropertyPhoto)
	adminGroup.GET("/hotels/:property_id/photos", inventoryHandler.ListPropertyPhotos)
	adminGroup.DELETE("/photos/property/:id", inventoryHandler.DeletePropertyPhoto)

	adminGroup.POST("/room-photos", inventoryHandler.AddRoomPhoto) // property_id/room_type_id/room_id via query params
	adminGroup.GET("/room-photos", inventoryHandler.ListRoomPhotos)
	adminGroup.DELETE("/photos/room/:id", inventoryHandler.DeleteRoomPhoto)

	// Admin user management
	adminGroup.POST("/users", adminHandler.CreateAdmin)
	adminGroup.GET("/users", adminHandler.ListAdmins)
	adminGroup.PUT("/users/:id", adminHandler.UpdateAdmin)
	adminGroup.POST("/users/:id/activate", adminHandler.Activate)
	adminGroup.POST("/users/:id/deactivate", adminHandler.Deactivate)

	// Booking oversight
	adminGroup.GET("/bookings", adminHandler.ListBookings)
	adminGroup.PUT("/bookings/:id/status", adminHandler.UpdateBookingStatus)

	// Reports
	adminGroup.GET("/reports/summary", reportHandler.Summary)
}
