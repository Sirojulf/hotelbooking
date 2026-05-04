package service

import (
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go/types"
)

type RegisterGuestInput struct {
	HotelID     string
	FullName    string
	Email       string
	Password    string
	PhoneNumber string
	Title       string
}

type GuestService interface {
	RegisterGuest(input RegisterGuestInput) (*models.Guest, error)
	LoginGuest(login, password string) (*types.TokenResponse, error)
	SearchHotels(query string) ([]models.Hotel, error)
	GetHotelDetails(hotelID string) (*models.HotelDetailResponse, error)
	GetMyReservations(guestID string) ([]models.Reservation, error)
	GetMyProfile(guestID string) (*models.Guest, error)
}

type guestService struct {
	guestRepo repository.GuestRepo
	hotelRepo repository.HotelRepo
	resRepo   repository.ReservationRepo
}

func NewGuestService(
	guestRepo repository.GuestRepo,
	hotelRepo repository.HotelRepo,
	resRepo repository.ReservationRepo,
) GuestService {
	return &guestService{
		guestRepo: guestRepo,
		hotelRepo: hotelRepo,
		resRepo:   resRepo,
	}
}

func (s *guestService) RegisterGuest(input RegisterGuestInput) (*models.Guest, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	if input.FullName == "" || input.Email == "" {
		return nil, fmt.Errorf("nama lengkap dan email wajib diisi")
	}

	hotelUUID, err := uuid.Parse(input.HotelID)
	if err != nil {
		return nil, fmt.Errorf("hotel_id tidak valid")
	}

	user, err := config.SupabaseClient.Auth.Signup(types.SignupRequest{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal mendaftar: %v", err)
	}

	guest := models.Guest{
		ID:          user.ID,
		HotelID:     hotelUUID,
		FullName:    input.FullName,
		Email:       user.Email,
		PhoneNumber: input.PhoneNumber,
		Title:       input.Title,
		LoyaltyTier: string(models.GuestTierBronze),
	}

	if err := s.guestRepo.CreateGuest(guest); err != nil {
		return nil, fmt.Errorf("gagal membuat profil tamu: %v", err)
	}

	return &guest, nil
}

func (s *guestService) LoginGuest(login, password string) (*types.TokenResponse, error) {
	if login == "" || password == "" {
		return nil, fmt.Errorf("email/phone dan password wajib diisi")
	}
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	isEmail := strings.Contains(login, "@")
	isPhone := isPhoneNumber(login)

	var (
		tokenResponse *types.TokenResponse
		err           error
	)
	switch {
	case isEmail:
		tokenResponse, err = config.SupabaseClient.Auth.SignInWithEmailPassword(login, password)
	case isPhone:
		tokenResponse, err = config.SupabaseClient.Auth.SignInWithPhonePassword(login, password)
	default:
		return nil, fmt.Errorf("format login tidak valid: gunakan email atau nomor telepon")
	}
	if err != nil {
		return nil, fmt.Errorf("login gagal: %v", err)
	}
	return tokenResponse, nil
}

func (s *guestService) SearchHotels(query string) ([]models.Hotel, error) {
	return s.hotelRepo.ListHotels(query)
}

func (s *guestService) GetHotelDetails(hotelID string) (*models.HotelDetailResponse, error) {
	hotel, err := s.hotelRepo.GetHotelByID(hotelID)
	if err != nil {
		return nil, err
	}
	roomTypes, err := s.hotelRepo.ListRoomTypes(hotelID)
	if err != nil {
		return nil, err
	}
	return &models.HotelDetailResponse{
		Hotel:     hotel,
		RoomTypes: roomTypes,
	}, nil
}

func (s *guestService) GetMyReservations(guestID string) ([]models.Reservation, error) {
	return s.resRepo.GetReservationsByGuestID(guestID)
}

func (s *guestService) GetMyProfile(guestID string) (*models.Guest, error) {
	return s.guestRepo.GetGuestByID(guestID)
}

func isPhoneNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) >= 10
}
