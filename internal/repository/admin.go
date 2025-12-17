package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

type AdminRepo interface {
	CreateAdmin(admin models.Admin) error
	GetAdminByEmail(email string) (*models.Admin, error)
	GetAdminByProperty(propertyID string) (*models.Admin, error)
	GetAdminByEmailAndProperty(email, propertyID string) (*models.Admin, error)
	UpdateActiveStatus(adminID string, isActive bool) error
}

type adminRepo struct{}

func NewAdminRepo() AdminRepo {
	return &adminRepo{}
}

func (r *adminRepo) CreateAdmin(admin models.Admin) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("admin").
		Insert(admin, false, "", "", "").
		Execute()

	if err != nil {
		return fmt.Errorf("gagal menambahkan admin ke database: %v", err)
	}
	return nil
}

func (r *adminRepo) GetAdminByEmail(email string) (*models.Admin, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("admin").
		Select("*", "", false).
		Eq("email", email).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil admin berdasarkan email: %v", err)
	}

	var admin models.Admin
	if err := json.Unmarshal(resp, &admin); err != nil {
		return nil, fmt.Errorf("gagal decode data admin: %v", err)
	}

	return &admin, nil
}

func (r *adminRepo) GetAdminByProperty(propertyID string) (*models.Admin, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("admin").
		Select("*", "", false).
		Eq("property_id", propertyID).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil admin berdasarkan property_id: %v", err)
	}

	var admin models.Admin
	if err := json.Unmarshal(resp, &admin); err != nil {
		return nil, fmt.Errorf("gagal decode data admin: %v", err)
	}

	return &admin, nil
}

func (r *adminRepo) GetAdminByEmailAndProperty(email, propertyID string) (*models.Admin, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("admin").
		Select("*", "", false).
		Eq("email", email).
		Eq("property_id", propertyID).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil admin berdasarkan email dan property_id: %v", err)
	}

	var admin models.Admin
	if err := json.Unmarshal(resp, &admin); err != nil {
		return nil, fmt.Errorf("gagal decode data admin: %v", err)
	}

	return &admin, nil
}

func (r *adminRepo) UpdateActiveStatus(adminID string, isActive bool) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	updateData := map[string]interface{}{
		"is_active": isActive,
	}

	_, _, err := config.SupabaseClient.
		From("admin").
		Update(updateData, "", "").
		Eq("id", adminID).
		Execute()

	if err != nil {
		return fmt.Errorf("gagal memperbarui status aktif admin: %v", err)
	}
	return nil
}
