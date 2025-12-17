package main

import (
	"hotelbooking/internal/config"
	"hotelbooking/internal/routes"

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
