package service

import (
	"fmt"
	"hotelbooking/internal/config"
	"strings"
	"unicode"

	"github.com/supabase-community/gotrue-go/types"
)

type LoginGuestService interface {
	Execute(login, password string) (*types.TokenResponse, error)
}

type loginGuestService struct{}

func NewLoginGuestService() LoginGuestService {
	return &loginGuestService{}
}

func (s *loginGuestService) Execute(login, password string) (*types.TokenResponse, error) {
	if login == "" || password == "" {
		return nil, fmt.Errorf("email/phone and password are required")
	}
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	var (
		tokenRespone *types.TokenResponse
		err          error
	)

	isEmail := strings.Contains(login, "@")
	isPhone := isNumeric(login)

	switch {
	case isEmail:
		tokenRespone, err = config.SupabaseClient.Auth.SignInWithEmailPassword(login, password)

	case isPhone:
		tokenRespone, err = config.SupabaseClient.Auth.SignInWithPhonePassword(login, password)

	default:
		return nil, fmt.Errorf("invalid login format: use email or phone number")

	}

	if err != nil {
		return nil, fmt.Errorf("login failed: %v", err)

	}
	return tokenRespone, nil
}

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) >= 12
}
