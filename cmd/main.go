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
	"net/http"
	"time"

	_ "hotelbooking/docs"
	"hotelbooking/internal/config"
	"hotelbooking/internal/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.HideBanner = true

	// ─── Middleware ───────────────────────────────────────────────────────────
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))
	// Logger hanya aktif di luar load test — comment saat benchmark
	e.Use(middleware.Logger())

	// ─── Health check (untuk k6 setup & readiness probe) ─────────────────────
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok", "time": time.Now().Unix()})
	})

	// ─── Supabase ─────────────────────────────────────────────────────────────
	if err := config.ConnectSupabase(); err != nil {
		e.Logger.Fatalf("failed to connect to supabase: %v", err)
	}

	// ─── Routes ───────────────────────────────────────────────────────────────
	routes.SetupRoutes(e)

	// ─── Server dengan timeout ────────────────────────────────────────────────
	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	e.Logger.Fatal(e.StartServer(s))
}
