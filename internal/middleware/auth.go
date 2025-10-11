package middleware

import (
	"hotelbooking/internal/config"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware memeriksa token JWT dari header Authorization.
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 1. Ambil header Authorization
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing authorization header"})
		}

		// 2. Periksa format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid authorization header format"})
		}
		token := parts[1]

		// 3. Validasi token ke Supabase (PROSES YANG BENAR)
		//    a. Buat klien baru dengan token yang diberikan
		authedClient := config.SupabaseClient.Auth.WithToken(token)
		//    b. Panggil GetUser() pada klien baru tersebut
		user, err := authedClient.GetUser()
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid or expired token"})
		}

		// 4. Simpan informasi pengguna di konteks untuk digunakan di handler
		c.Set("user", user)

		// 5. Jika valid, lanjutkan ke handler berikutnya
		return next(c)
	}
}
