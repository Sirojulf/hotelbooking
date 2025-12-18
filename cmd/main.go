package main

//go:generate swag init -g cmd/main.go -o docs
//
// @title Hotel Booking API
// @version 1.0
// @description REST API for hotel booking management (guest, booking, admin inventory, reports).
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"hotelbooking/internal/config"
	"hotelbooking/internal/routes"
	_ "hotelbooking/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Inisialisasi Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Hubungkan ke Supabase
	if err := config.ConnectSupabase(); err != nil {
		e.Logger.Fatalf("failed to connect to supabase: %v", err)
	}

	// Atur semua rute API
	routes.SetupRoutes(e)

	// Jalankan server di port 8080
	e.Logger.Fatal(e.Start(":8080"))

}
