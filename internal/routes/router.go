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
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hotel Booking API is running!")
	})

	apiV1 := e.Group("/api/v1")

	// Guest
	guestRepo := repository.NewGuestRepo()

	// Register
	registerSvc := service.NewRegisterGuestService(guestRepo)
	registerHandler := handler.NewRegisterGuestHandler(registerSvc)
	apiV1.POST("/guests/register", registerHandler.Handle)

	// Login
	loginSvc := service.NewLoginGuestService()
	loginHandler := handler.NewLoginGuestHandler(loginSvc)
	apiV1.POST("/guests/login", loginHandler.Handle)

	// Admin
	adminRepo := repository.NewAdminRepo()

	// Login
	adminLoginSvc := service.NewAdminLoginService(adminRepo)
	adminLoginHandler := handler.NewAdminLoginHandler(adminLoginSvc)
	apiV1.POST("/admin/login", adminLoginHandler.Login)

	// Protected Routes
	protectedRoutes := apiV1.Group("")
	protectedRoutes.Use(middleware.AuthMiddleware)

	// Inventory management
	propertyRepo := repository.NewPropertyRepo()
	inventorySvc := service.NewInventoryService(propertyRepo)
	inventoryHandler := handler.NewInventoryHandler(inventorySvc)

	// TODO: Nanti tambahkan middleware cek role Admin di sini!
	// Saat ini hanya cek token valid, tapi setidaknya route-nya sudah ada.

	protectedRoutes.POST("/admin/hotels", inventoryHandler.CreateHotel)

	// FIX: Route Room Types ditambahkan
	protectedRoutes.POST("/admin/room-types", inventoryHandler.CreateRoomType)

	// Add room
	protectedRoutes.POST("/admin/rooms", inventoryHandler.CreateRoom)

}
