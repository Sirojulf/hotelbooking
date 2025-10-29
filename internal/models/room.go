package models

type Room struct {
	ID                 string             `json:"id"`
	PropertyID         string             `json:"property_id"`
	RoomNumber         string             `json:"room_number"`
	RoomType           RoomType           `json:"room_type"`
	RateType           RateType           `json:"rate_type"`
	Status             RoomStatus         `json:"status"`
	HousekeepingStatus HousekeepingStatus `json:"housekeeping_status"`
	BasePrice          float64            `json:"base_price"`
	CreatedAt          string             `json:"created_at"`
}
