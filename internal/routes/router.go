package routes

import (
	"hotelbooking/internal/ai"
	"hotelbooking/internal/handler"
	"hotelbooking/internal/middleware"
	"hotelbooking/internal/repository"
	"hotelbooking/internal/service"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func SetupRoutes(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hotel Booking API is running!")
	})
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api/v1")

	// ======================
	// REPOSITORIES
	// ======================
	profileRepo := repository.NewProfileRepo()
	guestRepo := repository.NewGuestRepo()
	hotelRepo := repository.NewHotelRepo()
	reservationRepo := repository.NewReservationRepo()
	transactionRepo := repository.NewTransactionRepo()

	// ======================
	// SERVICES
	// ======================
	profileSvc := service.NewProfileService(profileRepo)
	guestSvc := service.NewGuestService(guestRepo, hotelRepo, reservationRepo)
	inventorySvc := service.NewInventoryService(hotelRepo)
	reservationSvc := service.NewReservationService(reservationRepo, hotelRepo, transactionRepo)
	reportSvc := service.NewReportService(reservationRepo, hotelRepo)
	aiSvc := service.NewAIService(hotelRepo)

	// ======================
	// HANDLERS
	// ======================
	// ─── AI Orchestrator (opsional — perlu OPENAI_API_KEY) ──────────────────
	aiOrchestrator, err := ai.NewOrchestrator(guestSvc, reservationSvc, hotelRepo)
	if err != nil {
		log.Printf("AI orchestrator nonaktif: %v (endpoint /ai/book akan return 503)", err)
		aiOrchestrator = nil
	} else {
		log.Println("AI orchestrator aktif (Function Calling agentic booking)")
	}

	adminHandler := handler.NewAdminHandler(profileSvc, reservationSvc)
	guestHandler := handler.NewGuestHandler(guestSvc)
	inventoryHandler := handler.NewInventoryHandler(inventorySvc)
	bookingHandler := handler.NewBookingHandler(reservationSvc)
	reportHandler := handler.NewReportHandler(reportSvc)
	aiHandler := handler.NewAIHandler(aiSvc, aiOrchestrator)

	// ======================
	// PUBLIC ROUTES
	// ======================
	api.POST("/auth/guest/register", guestHandler.Register)
	api.POST("/auth/guest/login", guestHandler.Login)
	api.POST("/auth/admin/login", adminHandler.Login)

	api.GET("/hotels", guestHandler.SearchHotels)
	api.GET("/hotels/:id", guestHandler.GetHotelDetail)
	api.GET("/rooms/:room_id/availability", bookingHandler.CheckAvailability)

	api.POST("/ai/recommend", aiHandler.RecommendRoom)

	// ======================
	// GUEST PROTECTED ROUTES
	// ======================
	guestGroup := api.Group("/guests")
	guestGroup.Use(middleware.AuthMiddleware)
	guestGroup.GET("/me", guestHandler.GetMyProfile)
	guestGroup.GET("/reservations", guestHandler.GetMyReservations)
	guestGroup.POST("/reservations", bookingHandler.CreateReservation)
	guestGroup.POST("/reservations/:id/pay", bookingHandler.PayReservation)
	guestGroup.POST("/reservations/:id/cancel", bookingHandler.CancelReservation)
	guestGroup.GET("/reservations/:id/transactions", bookingHandler.GetTransactions)
	guestGroup.POST("/ai/book", aiHandler.Book) // Agentic AI booking (Function Calling)

	// ======================
	// ADMIN PROTECTED ROUTES
	// ======================
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware)
	adminGroup.Use(middleware.AdminOnly(profileRepo))

	// Hotel management
	adminGroup.POST("/hotels", inventoryHandler.CreateHotel)
	adminGroup.GET("/hotels", inventoryHandler.ListHotels)
	adminGroup.PUT("/hotels/:id", inventoryHandler.UpdateHotel)
	adminGroup.DELETE("/hotels/:id", inventoryHandler.DeleteHotel)

	// Room type management
	adminGroup.POST("/room-types", inventoryHandler.CreateRoomType)
	adminGroup.GET("/room-types", inventoryHandler.ListRoomTypes)
	adminGroup.PUT("/room-types/:id", inventoryHandler.UpdateRoomType)
	adminGroup.DELETE("/room-types/:id", inventoryHandler.DeleteRoomType)

	// Room management
	adminGroup.POST("/rooms", inventoryHandler.CreateRoom)
	adminGroup.GET("/rooms", inventoryHandler.ListRooms)
	adminGroup.PUT("/rooms/:id", inventoryHandler.UpdateRoom)
	adminGroup.DELETE("/rooms/:id", inventoryHandler.DeleteRoom)

	// Staff management
	adminGroup.POST("/users", adminHandler.CreateProfile)
	adminGroup.GET("/users", adminHandler.ListProfiles)
	adminGroup.PUT("/users/:id", adminHandler.UpdateProfile)

	// Reservation management
	adminGroup.GET("/reservations", adminHandler.ListReservations)
	adminGroup.PUT("/reservations/:id", adminHandler.UpdateReservation)

	// Reports
	adminGroup.GET("/reports/summary", reportHandler.Summary)
}
