package service

import (
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go/types"
)

type CreateProfileInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role,omitempty"`
	HotelID  string `json:"hotel_id,omitempty"`
}

type ProfileLoginResponse struct {
	Profile *models.Profile      `json:"profile"`
	Session *types.TokenResponse `json:"session"`
}

type ProfileService interface {
	Login(email, password string) (*ProfileLoginResponse, error)
	CreateStaffProfile(input CreateProfileInput) (*models.Profile, error)
	UpdateRole(profileID, role string) error
	UpdateHotel(profileID, hotelID string) error
	ListProfiles(hotelID string) ([]models.Profile, error)
	GetProfileByID(id string) (*models.Profile, error)
	GetProfileByEmail(email string) (*models.Profile, error)
}

type profileService struct {
	repo repository.ProfileRepo
}

func NewProfileService(repo repository.ProfileRepo) ProfileService {
	return &profileService{repo: repo}
}

func (s *profileService) Login(email, password string) (*ProfileLoginResponse, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email dan password wajib diisi")
	}
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	session, err := config.SupabaseClient.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("login gagal: %v", err)
	}

	profile, err := s.repo.GetProfileByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("profil tidak ditemukan: %v", err)
	}

	return &ProfileLoginResponse{
		Profile: profile,
		Session: session,
	}, nil
}

func (s *profileService) CreateStaffProfile(input CreateProfileInput) (*models.Profile, error) {
	if input.Email == "" || input.Password == "" {
		return nil, fmt.Errorf("email dan password wajib diisi")
	}
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	user, err := config.SupabaseClient.Auth.Signup(types.SignupRequest{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan user: %v", err)
	}

	profile := models.Profile{
		ID:       user.ID,
		Email:    user.Email,
		FullName: input.FullName,
		Role:     input.Role,
	}

	if input.HotelID != "" {
		hotelUUID, err := uuid.Parse(input.HotelID)
		if err != nil {
			return nil, fmt.Errorf("hotel_id tidak valid")
		}
		profile.HotelID = &hotelUUID
	}

	if err := s.repo.CreateProfile(profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (s *profileService) UpdateRole(profileID, role string) error {
	return s.repo.UpdateProfileRole(profileID, role)
}

func (s *profileService) UpdateHotel(profileID, hotelID string) error {
	return s.repo.UpdateProfileHotel(profileID, hotelID)
}

func (s *profileService) ListProfiles(hotelID string) ([]models.Profile, error) {
	return s.repo.ListProfiles(hotelID)
}

func (s *profileService) GetProfileByID(id string) (*models.Profile, error) {
	return s.repo.GetProfileByID(id)
}

func (s *profileService) GetProfileByEmail(email string) (*models.Profile, error) {
	return s.repo.GetProfileByEmail(email)
}
