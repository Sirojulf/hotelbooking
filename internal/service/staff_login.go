package service

import (
	"fmt"
	"hotelbooking/internal/config"

	"github.com/supabase-community/gotrue-go/types"
)

type LoginStaffService interface {
	Execute(email, password string) (*types.TokenResponse, error)
}

type loginStaffService struct{}

func NewLoginStaffService() LoginStaffService {
	return &loginStaffService{}
}

func (s *loginStaffService) Execute(email, password string) (*types.TokenResponse, error) {
	tokenRespone, err := config.SupabaseClient.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("email atau password salah: %v", err)

	}
	return tokenRespone, nil

}
