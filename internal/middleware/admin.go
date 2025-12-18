package middleware

import (
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supabase-community/gotrue-go/types"
)

func AdminOnly(adminRepo repository.AdminRepo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*types.User)
			if !ok || user == nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
			}

			admin, err := adminRepo.GetAdminByID(user.ID.String())
			if err != nil {
				return c.JSON(http.StatusForbidden, echo.Map{"error": "Admin access required"})
			}
			if !admin.IsActive {
				return c.JSON(http.StatusForbidden, echo.Map{"error": "Admin account is inactive"})
			}

			if role := extractRole(user); role != "" {
				admin.Role = role
			}

			c.Set("admin", admin)
			return next(c)
		}
	}
}

func GetAdminFromContext(c echo.Context) (*models.Admin, bool) {
	admin, ok := c.Get("admin").(*models.Admin)
	return admin, ok && admin != nil
}

func extractRole(user *types.User) string {
	if user == nil {
		return ""
	}
	if role := extractRoleFromMap(user.AppMetadata); role != "" {
		return role
	}
	return extractRoleFromMap(user.UserMetadata)
}

func extractRoleFromMap(data map[string]any) string {
	if data == nil {
		return ""
	}
	role, _ := data["role"].(string)
	return role
}
