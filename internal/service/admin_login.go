package service

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"

	"github.com/supabase-community/gotrue-go/types"
)

type AdminLoginService interface {
	Execute(email, password, propertyID string) (*AdminLoginResponse, error)
}

type AdminLoginResponse struct {
	Admin   *models.Admin        `json:"admin"`
	Session *types.TokenResponse `json:"session"`
}

type adminLoginService struct {
	repo repository.AdminRepo
}

func NewAdminLoginService(repo repository.AdminRepo) AdminLoginService {
	return &adminLoginService{repo: repo}
}

func (s *adminLoginService) Execute(email, password, propertyID string) (*AdminLoginResponse, error) {
	if email == "" || password == "" || propertyID == "" {
		return nil, fmt.Errorf("email, password, and property_id are required")
	}

	tokenResponse, err := config.SupabaseClient.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}

	resp, _, err := config.SupabaseClient.
		From("admin").
		Select("*", "", false).
		Eq("property_id", propertyID).
		Eq("email", email).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("admin not found or property mismatch: %v", err)
	}

	var admin models.Admin
	if err := json.Unmarshal(resp, &admin); err != nil {
		return nil, fmt.Errorf("failed to parse admin data: %v", err)
	}

	if !admin.IsActive {
		return nil, fmt.Errorf("admin account is inactive")
	}

	return &AdminLoginResponse{
		Admin:   &admin,
		Session: tokenResponse,
	}, nil
}
