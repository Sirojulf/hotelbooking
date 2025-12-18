package handler

import "hotelbooking/internal/models"

type TokenResponseDoc struct {
	AccessToken  string                 `json:"access_token"`
	RefreshToken string                 `json:"refresh_token"`
	TokenType    string                 `json:"token_type"`
	ExpiresIn    int                    `json:"expires_in"`
	ExpiresAt    int64                  `json:"expires_at"`
	User         map[string]interface{} `json:"user"`
}

type AdminLoginResponseDoc struct {
	Admin   *models.Admin    `json:"admin"`
	Session TokenResponseDoc `json:"session"`
}

type RoomRateDoc struct {
	ID               string          `json:"id"`
	RoomID           string          `json:"room_id"`
	Date             string          `json:"date"`
	AvailableRooms   int             `json:"available_rooms"`
	LinearRate       *float64        `json:"linear_rate"`
	NonLinearRate    interface{}     `json:"non_linear_rate"`
	MinNights        int             `json:"min_nights"`
	MaxNights        int             `json:"max_nights"`
	StopSell         bool            `json:"stop_sell"`
	CloseOnArrival   bool            `json:"close_on_arrival"`
	CloseOnDeparture bool            `json:"close_on_departure"`
	CreatedAt        string          `json:"created_at"`
}

type RoomRateRequestDoc struct {
	RoomID           string      `json:"room_id"`
	Dates            []string    `json:"dates"`
	AvailableRooms   int         `json:"available_rooms"`
	LinearRate       *float64    `json:"linear_rate"`
	NonLinearRate    interface{} `json:"non_linear_rate"`
	MinNights        int         `json:"min_nights"`
	MaxNights        int         `json:"max_nights"`
	StopSell         bool        `json:"stop_sell"`
	CloseOnArrival   bool        `json:"close_on_arrival"`
	CloseOnDeparture bool        `json:"close_on_departure"`
}
