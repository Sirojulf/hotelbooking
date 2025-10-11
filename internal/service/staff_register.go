package service

import (
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"

	"github.com/supabase-community/gotrue-go/types"
)

type RegisterStaffService interface {
	Execute(email, password, fullname, role string) (*models.StaffProfile, error)
}

type registerStaffService struct {
	staffRepo repository.StaffRepo
}

func NewRegisterStaffService(staffRepo repository.StaffRepo) RegisterStaffService {
	return &registerStaffService{staffRepo: staffRepo}
}

func (s *registerStaffService) Execute(email, password, fullname, role string) (*models.StaffProfile, error) {
	userMetadata := map[string]interface{}{
		"user_role": role,
	}

	requestBody := types.SignupRequest{
		Email:    email,
		Password: password,
		Data:     userMetadata,
	}

	user, err := config.SupabaseClient.Auth.Signup(requestBody)
	if err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan staf: %v", err)

	}

	staffProfile := models.StaffProfile{
		ID:       user.ID,
		FullName: fullname,
		Email:    user.Email,
		Role:     role,
	}

	if err := s.staffRepo.CreateProfile(staffProfile); err != nil {
		return nil, fmt.Errorf("gagal membuat profil staf: %v", err)
	}

	return &staffProfile, nil

}
