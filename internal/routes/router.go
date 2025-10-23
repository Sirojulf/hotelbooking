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

	// AuthMiddleware
	protectedRoutes := apiV1.Group("")
	protectedRoutes.Use(middleware.AuthMiddleware)

	// Hotel
	hotelRepo := repository.NewHotelRepo()

	// create_hotel
	createHotelSvc := service.NewCreateHotelService(hotelRepo)
	createHotelHandler := handler.NewCreateHotelHandler(createHotelSvc)
	protectedRoutes.POST("/hotels/create", createHotelHandler.Handle)

}
