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
	"hotelbooking/internal/repository"
	"hotelbooking/internal/routes"

	json "github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// fastJSONSerializer = serializer Echo pakai goccy/go-json (2-3× lebih cepat dari encoding/json)
type fastJSONSerializer struct{}

func (fastJSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := json.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

func (fastJSONSerializer) Deserialize(c echo.Context, i interface{}) error {
	return json.NewDecoder(c.Request().Body).Decode(i)
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.JSONSerializer = fastJSONSerializer{}

	// ─── Middleware ───────────────────────────────────────────────────────────
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level:     1,    // BestSpeed — minim CPU
		MinLength: 1024, // skip kompresi response < 1 KB
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))
	// Logger hanya aktif di luar load test — comment saat benchmark
	// e.Use(middleware.Logger())

	// ─── Health check (untuk k6 setup & readiness probe) ─────────────────────
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok", "time": time.Now().Unix()})
	})

	// ─── Supabase ─────────────────────────────────────────────────────────────
	if err := config.ConnectSupabase(); err != nil {
		e.Logger.Fatalf("failed to connect to supabase: %v", err)
	}

	// ─── Pre-warm cache room & room_type ──────────────────────────────────────
	// Hilangkan cache-miss di request pertama; rooms/room_types jarang berubah.
	repository.WarmRoomCache()

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
