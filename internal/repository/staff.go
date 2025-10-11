package repository

import (
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

type StaffRepo interface {
	CreateProfile(profile models.StaffProfile) error
}

type staffRepo struct{}

func NewStaffRepo() StaffRepo {
	return &staffRepo{}
}

func (r *staffRepo) CreateProfile(profile models.StaffProfile) error {
	_, _, err := config.SupabaseClient.From("staff_profiles").Insert(profile, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("gagal menyisipkan profil staf ke db: %v", err)
	}
	return nil
}
