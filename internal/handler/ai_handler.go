package handler

import (
	"hotelbooking/internal/ai"
	"hotelbooking/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supabase-community/gotrue-go/types"
)

type AIHandler struct {
	svc          service.AIService
	orchestrator *ai.Orchestrator // optional — nil kalau OPENAI_API_KEY tidak diset
}

func NewAIHandler(aiSvc service.AIService, orchestrator *ai.Orchestrator) *AIHandler {
	return &AIHandler{svc: aiSvc, orchestrator: orchestrator}
}

type AIBookRequest struct {
	Message string `json:"message"`
}

// POST /api/v1/guests/ai/book
// @Summary Booking otomatis via natural language (Agentic AI)
// @Description AI agent (Function Calling) yang otomatis cari hotel, cek availability, create reservation, dan bayar. Butuh OPENAI_API_KEY di .env.
// @Tags AI
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body AIBookRequest true "Pesan natural language user"
// @Success 200 {object} ai.Result
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /guests/ai/book [post]
func (h *AIHandler) Book(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, echo.Map{
			"error": "AI booking belum aktif — set OPENAI_API_KEY di .env lalu restart server",
		})
	}
	user, ok := c.Get("user").(*types.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	var req AIBookRequest
	if err := c.Bind(&req); err != nil || req.Message == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "field 'message' wajib diisi"})
	}

	result, err := h.orchestrator.RunBooking(c.Request().Context(), user.ID.String(), req.Message)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
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
