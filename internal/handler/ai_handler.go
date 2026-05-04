package handler

import (
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AIHandler struct {
	svc service.AIService
}

func NewAIHandler(aiSvc service.AIService) *AIHandler {
	return &AIHandler{svc: aiSvc}
}

// POST /api/v1/ai/recommend
// @Summary Rekomendasi kamar berbasis AI
// @Description Agentic AI memilihkan kamar terbaik berdasarkan preferensi tamu. Set AI_MOCK=true di .env untuk load testing skala besar.
// @Tags AI
// @Accept json
// @Produce json
// @Param payload body service.AIRecommendRequest true "Preferensi tamu"
// @Success 200 {object} service.AIRecommendResult
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /ai/recommend [post]
func (h *AIHandler) RecommendRoom(c echo.Context) error {
	var req service.AIRecommendRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Request tidak valid"})
	}
	if req.Query == "" || req.CheckIn == "" || req.CheckOut == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "query, check_in, dan check_out wajib diisi",
		})
	}

	result, err := h.svc.RecommendRoom(req)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}
