package service

import (
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go/types"
)

// =====================
// DTO / Types
// =====================

// Dipakai saat membuat admin baru
type CreateAdminInput struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	PropertyID string `json:"property_id,omitempty"`
	Role       string `json:"role,omitempty"`
}

// Response saat admin login
type AdminLoginResponse struct {
	Admin   *models.Admin        `json:"admin"`
	Session *types.TokenResponse `json:"session"`
}

// =====================
// Interface AdminService
// =====================

type AdminService interface {
	// AUTH
	Login(email, password string) (*AdminLoginResponse, error)

	// MANAGEMENT
	CreateAdmin(input CreateAdminInput) (*models.Admin, error)
	ActivateAdmin(adminID string) error
	DeactivateAdmin(adminID string) error
	UpdateRole(adminID, role string) error
	UpdateProperty(adminID, propertyID string) error
	ListAdmins(propertyID string) ([]models.Admin, error)

	// QUERY
	GetAdminByEmail(email string) (*models.Admin, error)
	GetAdminByID(id string) (*models.Admin, error)
	GetAdminForProperty(propertyID string) (*models.Admin, error)
}

// =====================
// Implementasi
// =====================

type adminService struct {
	repo repository.AdminRepo
}

func NewAdminService(repo repository.AdminRepo) AdminService {
	return &adminService{repo: repo}
}

// ----------------------
// AUTH
// ----------------------

// Login admin: cek admin exist & aktif, lalu login via Supabase Auth
func (s *adminService) Login(email, password string) (*AdminLoginResponse, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password are required")
	}
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	// 1. Pastikan email ini memang admin dan aktif
	admin, err := s.repo.GetAdminByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("admin tidak terdaftar: %v", err)
	}
	if !admin.IsActive {
		return nil, fmt.Errorf("admin account is inactive")
	}

	// 2. Validasi kredensial ke Supabase Auth
	session, err := config.SupabaseClient.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}

	return &AdminLoginResponse{
		Admin:   admin,
		Session: session,
	}, nil
}

// ----------------------
// MANAGEMENT
// ----------------------

// CreateAdmin: buat user di Supabase Auth + row di tabel admin
func (s *adminService) CreateAdmin(input CreateAdminInput) (*models.Admin, error) {
	if input.Email == "" || input.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	// 1. Signup ke Supabase Auth
	req := types.SignupRequest{
		Email:    input.Email,
		Password: input.Password,
	}
	user, err := config.SupabaseClient.Auth.Signup(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan admin: %v", err)
	}

	// 2. Buat row di tabel admin
	var propertyUUID *uuid.UUID
	if input.PropertyID != "" {
		parsed, err := uuid.Parse(input.PropertyID)
		if err != nil {
			return nil, fmt.Errorf("invalid property id")
		}
		propertyUUID = &parsed
	}

	admin := models.Admin{
		ID:         user.ID,    // id disamakan dengan auth.users
		Email:      user.Email, // pakai email dari Supabase
		Role:       input.Role,
		IsActive:   true,
		PropertyID: propertyUUID,
	}

	if err := s.repo.CreateAdmin(admin); err != nil {
		return nil, err
	}

	return &admin, nil
}

func (s *adminService) ActivateAdmin(adminID string) error {
	return s.repo.UpdateActiveStatus(adminID, true)
}

func (s *adminService) DeactivateAdmin(adminID string) error {
	return s.repo.UpdateActiveStatus(adminID, false)
}

func (s *adminService) UpdateRole(adminID, role string) error {
	if role == "" {
		return fmt.Errorf("role is required")
	}
	return s.repo.UpdateRole(adminID, role)
}

func (s *adminService) UpdateProperty(adminID, propertyID string) error {
	if propertyID == "" {
		return fmt.Errorf("property_id is required")
	}
	return s.repo.UpdateProperty(adminID, propertyID)
}

func (s *adminService) ListAdmins(propertyID string) ([]models.Admin, error) {
	return s.repo.ListAdmins(propertyID)
}

// ----------------------
// QUERY
// ----------------------

func (s *adminService) GetAdminByEmail(email string) (*models.Admin, error) {
	return s.repo.GetAdminByEmail(email)
}

func (s *adminService) GetAdminByID(id string) (*models.Admin, error) {
	return s.repo.GetAdminByID(id)
}

func (s *adminService) GetAdminForProperty(propertyID string) (*models.Admin, error) {
	return s.repo.GetAdminByProperty(propertyID)
}
