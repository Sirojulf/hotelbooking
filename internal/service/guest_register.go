package service

import (
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"

	"github.com/supabase-community/gotrue-go/types" // <-- IMPORT BARU
)

type RegisterGuestService interface {
	Execute(email, password, firstName, lastName, phone string, gender models.Gender) (*models.Guest, error)
}

type registerGuestService struct {
	guestRepo repository.GuestRepo
}

func NewRegisterGuestService(guestRepo repository.GuestRepo) *registerGuestService {
	return &registerGuestService{guestRepo: guestRepo}
}

func (s *registerGuestService) Execute(email, password, firstName, lastName, phone string, gender models.Gender) (*models.Guest, error) {
	// PERBAIKAN: Buat struct SignupRequest terlebih dahulu
	requestBody := types.SignupRequest{
		Email:    email,
		Password: password,
		// Anda juga bisa menambahkan data lain di sini jika diperlukan
		// Data: map[string]interface{}{"full_name": firstName + " " + lastName},
	}

	// Panggil fungsi Signup dengan satu argumen struct
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
		GuestType: models.GuestTypedult,
		VIPStatus: models.VIPStatusBronze,
		Gender:    gender,
	}

	if err := s.guestRepo.CreateProfile(guestProfile); err != nil {
		return nil, fmt.Errorf("gagal membuat profil: %v", err)
	}

	return &guestProfile, nil
}
