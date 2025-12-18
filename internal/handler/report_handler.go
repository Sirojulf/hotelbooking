package handler

import (
	"hotelbooking/internal/middleware"
	"hotelbooking/internal/service"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type ReportHandler struct {
	Svc service.ReportService
}

func NewReportHandler(svc service.ReportService) *ReportHandler {
	return &ReportHandler{Svc: svc}
}

// @Summary Get summary report
// @Tags Reports
// @Security BearerAuth
// @Produce json
// @Param property_id query string false "Property ID"
// @Param start query string true "Start date (YYYY-MM-DD)"
// @Param end query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} service.ReportSummary
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/reports/summary [get]
func (h *ReportHandler) Summary(c echo.Context) error {
	admin, ok := middleware.GetAdminFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	propertyID := c.QueryParam("property_id")
	if admin.PropertyID != nil {
		propertyID = admin.PropertyID.String()
	}
	start := c.QueryParam("start")
	end := c.QueryParam("end")
	if start == "" || end == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "start and end are required"})
	}

	var startTime, endTime time.Time
	var err error
	if start != "" {
		startTime, err = time.Parse("2006-01-02", start)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid start date"})
		}
	}
	if end != "" {
		endTime, err = time.Parse("2006-01-02", end)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid end date"})
		}
	}

	summary, err := h.Svc.GetSummary(propertyID, startTime, endTime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, summary)
}
