package middleware

import (
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supabase-community/gotrue-go/types"
)

func AdminOnly(profileRepo repository.ProfileRepo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*types.User)
			if !ok || user == nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
			}

			profile, err := profileRepo.GetProfileByID(user.ID.String())
			if err != nil {
				return c.JSON(http.StatusForbidden, echo.Map{"error": "Akses admin diperlukan"})
			}

			c.Set("profile", profile)
			return next(c)
		}
	}
}

func GetProfileFromContext(c echo.Context) (*models.Profile, bool) {
	profile, ok := c.Get("profile").(*models.Profile)
	return profile, ok && profile != nil
}

func IsSuperAdmin(profile *models.Profile) bool {
	return profile != nil && profile.HotelID == nil
}
