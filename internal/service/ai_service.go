package service

import (
	"fmt"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"

	"github.com/spf13/viper"
)

type AIRecommendRequest struct {
	Query    string  `json:"query"`
	CheckIn  string  `json:"check_in"`
	CheckOut string  `json:"check_out"`
	Budget   float64 `json:"budget,omitempty"`
	Prefs    string  `json:"preferences,omitempty"`
}

type AIRecommendResult struct {
	RecommendedRoomID string  `json:"recommended_room_id"`
	HotelID           string  `json:"hotel_id"`
	HotelName         string  `json:"hotel_name"`
	RoomNumber        string  `json:"room_number"`
	EstimatedPrice    float64 `json:"estimated_price"`
	Reasoning         string  `json:"reasoning"`
	PromoMessage      string  `json:"promo_message,omitempty"`
}

type AIService interface {
	RecommendRoom(req AIRecommendRequest) (*AIRecommendResult, error)
}

type aiService struct {
	hotelRepo repository.HotelRepo
}

func NewAIService(hotelRepo repository.HotelRepo) AIService {
	return &aiService{hotelRepo: hotelRepo}
}

func (s *aiService) RecommendRoom(req AIRecommendRequest) (*AIRecommendResult, error) {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()
	if viper.GetBool("AI_MOCK") {
		return mockAIRecommendation(), nil
	}

	hotels, err := s.hotelRepo.ListHotels(req.Query)
	if err != nil || len(hotels) == 0 {
		return nil, fmt.Errorf("tidak ada hotel tersedia untuk pencarian '%s'", req.Query)
	}

	var bestRoom *models.Room
	var bestHotel *models.Hotel
	var bestPrice float64

	for i := range hotels {
		hotel := &hotels[i]
		rooms, err := s.hotelRepo.ListRooms(hotel.ID.String(), "")
		if err != nil {
			continue
		}
		for j := range rooms {
			room := &rooms[j]
			if room.Status != models.RoomStatusAvailable {
				continue
			}
			rt, err := s.hotelRepo.GetRoomTypeByID(room.RoomTypeID.String())
			if err != nil {
				continue
			}
			price := rt.PricePerNight
			if price == 0 {
				price = rt.BasePrice
			}
			if req.Budget > 0 && price > req.Budget {
				continue
			}
			if bestRoom == nil || price < bestPrice {
				bestRoom = room
				bestHotel = hotel
				bestPrice = price
			}
		}
	}

	if bestRoom == nil {
		return nil, fmt.Errorf("tidak ada kamar tersedia sesuai kriteria")
	}

	return &AIRecommendResult{
		RecommendedRoomID: bestRoom.ID.String(),
		HotelID:           bestHotel.ID.String(),
		HotelName:         bestHotel.Name,
		RoomNumber:        bestRoom.RoomNumber,
		EstimatedPrice:    bestPrice,
		Reasoning:         "Kamar terjangkau yang tersedia sesuai kriteria pencarian.",
		PromoMessage:      "",
	}, nil
}

func mockAIRecommendation() *AIRecommendResult {
	return &AIRecommendResult{
		RecommendedRoomID: "00000000-0000-0000-0000-000000000001",
		HotelID:           "00000000-0000-0000-0000-000000000002",
		HotelName:         "Hotel Mock",
		RoomNumber:        "101",
		EstimatedPrice:    500000,
		Reasoning:         "Kamar ini sesuai dengan preferensi budget dan lokasi.",
		PromoMessage:      "Diskon 10% untuk pemesanan lebih dari 3 malam!",
	}
}
