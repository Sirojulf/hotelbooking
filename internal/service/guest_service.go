package service

import (
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"strings"
	"unicode"

	"github.com/supabase-community/gotrue-go/types"
)

type RegisterGuestInput struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	Phone     string
	Gender    models.Gender
	Country   string
}

type GuestService interface {
	RegisterGuest(input RegisterGuestInput) (*models.Guest, error)
	LoginGuest(login, password string) (*types.TokenResponse, error)
	SearchHotels(city string) ([]models.Properties, error)
	GetHotelDetails(propertyID string) (*models.PropertyDetailResponse, error)
	GetMyBookings(guestID string) ([]models.Booking, error)
	GetMyProfile(guestID string) (*models.Guest, error)
}

type guestService struct {
	guestRepo repository.GuestRepo
	propRepo  repository.PropertyRepo
	bookRepo  repository.BookingRepo
}

func NewGuestService(
	guestRepo repository.GuestRepo,
	propRepo repository.PropertyRepo,
	bookRepo repository.BookingRepo,
) GuestService {
	return &guestService{
		guestRepo: guestRepo,
		propRepo:  propRepo,
		bookRepo:  bookRepo,
	}
}

// ----------------- AUTH ---------------------

func (s *guestService) RegisterGuest(input RegisterGuestInput) (*models.Guest, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	req := types.SignupRequest{
		Email:    input.Email,
		Password: input.Password,
	}

	user, err := config.SupabaseClient.Auth.Signup(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mendaftar: %v", err)
	}

	guestProfile := models.Guest{
		ID:        user.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     user.Email,
		Phone:     input.Phone,
		GuestType: models.GuestTypeAdult,
		VIPStatus: models.VIPStatusBronze,
		Gender:    input.Gender,
		Country:   input.Country,
	}

	if err := s.guestRepo.CreateProfile(guestProfile); err != nil {
		return nil, fmt.Errorf("gagal membuat profil: %v", err)
	}

	return &guestProfile, nil
}

func (s *guestService) LoginGuest(login, password string) (*types.TokenResponse, error) {
	if login == "" || password == "" {
		return nil, fmt.Errorf("email/phone and password are required")
	}
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	var (
		tokenResponse *types.TokenResponse
		err           error
	)

	isEmail := strings.Contains(login, "@")
	isPhone := isNumeric(login)

	switch {
	case isEmail:
		tokenResponse, err = config.SupabaseClient.Auth.SignInWithEmailPassword(login, password)
	case isPhone:
		tokenResponse, err = config.SupabaseClient.Auth.SignInWithPhonePassword(login, password)
	default:
		return nil, fmt.Errorf("invalid login format: use email or phone number")
	}

	if err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}
	return tokenResponse, nil
}

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) >= 12
}

// --------------- EXPERIENCE -----------------

func (s *guestService) SearchHotels(city string) ([]models.Properties, error) {
	return s.propRepo.SearchProperties(city)
}

func (s *guestService) GetHotelDetails(propertyID string) (*models.PropertyDetailResponse, error) {
	property, err := s.propRepo.GetPropertyByID(propertyID)
	if err != nil {
		return nil, err
	}

	roomTypes, err := s.propRepo.GetRoomTypesByPropertyID(propertyID)
	if err != nil {
		return nil, err
	}

	return &models.PropertyDetailResponse{
		Property:  property,
		RoomTypes: roomTypes,
	}, nil
}

func (s *guestService) GetMyBookings(guestID string) ([]models.Booking, error) {
	return s.bookRepo.GetBookingsByGuestID(guestID)
}

func (s *guestService) GetMyProfile(guestID string) (*models.Guest, error) {
	return s.guestRepo.GetGuestByID(guestID)
}
