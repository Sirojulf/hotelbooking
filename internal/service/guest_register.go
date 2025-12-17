package service

import (
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"

	"github.com/supabase-community/gotrue-go/types"
)

type RegisterGuestService interface {
	Execute(email, password, firstName, lastName, phone, country string, gender models.Gender) (*models.Guest, error)
}

type registerGuestService struct {
	guestRepo repository.GuestRepo
}

func NewRegisterGuestService(guestRepo repository.GuestRepo) *registerGuestService {
	return &registerGuestService{guestRepo: guestRepo}
}

func (s *registerGuestService) Execute(email, password, firstName, lastName, phone, country string, gender models.Gender) (*models.Guest, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	requestBody := types.SignupRequest{
		Email:    email,
		Password: password,
	}

	user, err := config.SupabaseClient.Auth.Signup(requestBody)
	if err != nil {
		return nil, fmt.Errorf("gagal mendaftar: %v", err)
	}

	guestProfile := models.Guest{
		ID:        user.ID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     user.Email,
		Phone:     phone,
		GuestType: models.GuestTypeAdult,
		VIPStatus: models.VIPStatusBronze,
		Gender:    gender,
		Country:   country,
	}

	if err := s.guestRepo.CreateProfile(guestProfile); err != nil {
		return nil, fmt.Errorf("gagal membuat profil: %v", err)
	}

	return &guestProfile, nil
}
