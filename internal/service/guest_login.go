package service

import (
	"fmt"
	"hotelbooking/internal/config"

	"github.com/supabase-community/gotrue-go/types"
)

type LoginGuestService interface {
	Execute(email, password string) (*types.TokenResponse, error)
}

type loginGuestService struct{}

func NewLoginGuestService() LoginGuestService {
	return &loginGuestService{}
}

func (s *loginGuestService) Execute(email, password string) (*types.TokenResponse, error) {
	// Panggil fungsi SignInWithEmailPassword, yang mengembalikan *types.TokenResponse
	tokenResponse, err := config.SupabaseClient.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("email atau password salah: %v", err)
	}

	// PERBAIKAN: Kembalikan pointer 'tokenResponse' secara langsung,
	// tanpa menggunakan '&'.
	return tokenResponse, nil
}
