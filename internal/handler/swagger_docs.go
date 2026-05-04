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
	Profile *models.Profile  `json:"profile"`
	Session TokenResponseDoc `json:"session"`
}
