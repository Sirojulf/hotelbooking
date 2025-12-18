// internal/repository/admin.go
package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

const adminTable = "admin"

type AdminRepo interface {
	CreateAdmin(admin models.Admin) error
	GetAdminByEmail(email string) (*models.Admin, error)
	GetAdminByID(id string) (*models.Admin, error)
	GetAdminByProperty(propertyID string) (*models.Admin, error)
	GetAdminByEmailAndProperty(email, propertyID string) (*models.Admin, error)
	UpdateActiveStatus(adminID string, isActive bool) error
	UpdateRole(adminID, role string) error
	UpdateProperty(adminID, propertyID string) error
	ListAdmins(propertyID string) ([]models.Admin, error)
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
		From(adminTable).
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
		From(adminTable).
		Select("*", "", false).
		Eq("email", email).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("admin tidak ditemukan: %v", err)
	}

	var admin models.Admin
	if err := json.Unmarshal(resp, &admin); err != nil {
		return nil, fmt.Errorf("gagal decode data admin: %v", err)
	}

	return &admin, nil
}

func (r *adminRepo) GetAdminByID(id string) (*models.Admin, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}

	resp, _, err := config.SupabaseClient.
		From(adminTable).
		Select("*", "", false).
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("admin tidak ditemukan: %v", err)
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
		From(adminTable).
		Select("*", "", false).
		Eq("property_id", propertyID).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("admin berdasarkan property_id tidak ditemukan: %v", err)
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
		From(adminTable).
		Select("*", "", false).
		Eq("email", email).
		Eq("property_id", propertyID).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("admin tidak ditemukan untuk email dan property_id tersebut: %v", err)
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

	updateData := map[string]any{
		"is_active": isActive,
	}

	_, _, err := config.SupabaseClient.
		From(adminTable).
		Update(updateData, "", "").
		Eq("id", adminID).
		Execute()
	if err != nil {
		return fmt.Errorf("gagal memperbarui status aktif admin: %v", err)
	}
	return nil
}

func (r *adminRepo) UpdateRole(adminID, role string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	data := map[string]any{"role": role}
	_, _, err := config.SupabaseClient.
		From(adminTable).
		Update(data, "", "").
		Eq("id", adminID).
		Execute()
	if err != nil {
		return fmt.Errorf("gagal memperbarui role admin: %v", err)
	}
	return nil
}

func (r *adminRepo) UpdateProperty(adminID, propertyID string) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	data := map[string]any{"property_id": propertyID}
	_, _, err := config.SupabaseClient.
		From(adminTable).
		Update(data, "", "").
		Eq("id", adminID).
		Execute()
	if err != nil {
		return fmt.Errorf("gagal memperbarui property admin: %v", err)
	}
	return nil
}

func (r *adminRepo) ListAdmins(propertyID string) ([]models.Admin, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	q := config.SupabaseClient.
		From(adminTable).
		Select("*", "", false)
	if propertyID != "" {
		q = q.Eq("property_id", propertyID)
	}
	resp, _, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar admin: %v", err)
	}
	var admins []models.Admin
	if err := json.Unmarshal(resp, &admins); err != nil {
		return nil, fmt.Errorf("gagal decode data admin: %v", err)
	}
	return admins, nil
}
