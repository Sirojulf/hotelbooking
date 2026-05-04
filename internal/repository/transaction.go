package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
)

type TransactionRepo interface {
	CreateTransaction(tx models.Transaction) error
	GetTransactionsByReservationID(reservationID string) ([]models.Transaction, error)
	ListTransactionsByHotel(hotelID string) ([]models.Transaction, error)
}

type transactionRepo struct{}

func NewTransactionRepo() TransactionRepo {
	return &transactionRepo{}
}

func (r *transactionRepo) CreateTransaction(tx models.Transaction) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.From("transactions").Insert(tx, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("gagal membuat transaksi: %v", err)
	}
	return nil
}

func (r *transactionRepo) GetTransactionsByReservationID(reservationID string) ([]models.Transaction, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("transactions").
		Select("*", "", false).
		Eq("reservation_id", reservationID).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil transaksi: %v", err)
	}
	var txs []models.Transaction
	if err := json.Unmarshal(resp, &txs); err != nil {
		return nil, err
	}
	return txs, nil
}

func (r *transactionRepo) ListTransactionsByHotel(hotelID string) ([]models.Transaction, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("transactions").
		Select("*, reservations!inner(hotel_id)", "", false).
		Eq("reservations.hotel_id", hotelID).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil transaksi hotel: %v", err)
	}
	var txs []models.Transaction
	if err := json.Unmarshal(resp, &txs); err != nil {
		return nil, err
	}
	return txs, nil
}
