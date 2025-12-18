package repository

import (
	"encoding/json"
	"fmt"
	"hotelbooking/internal/config"
	"hotelbooking/internal/models"
	"time"
)

type PaymentRepo interface {
	CreatePayment(payment models.Payment) error
	GetPaymentByBookingID(bookingID string) (*models.Payment, error)
	UpdatePaymentStatus(bookingID string, status models.PaymentStatus, provider, reference string) (*models.Payment, error)
	CreateInvoice(invoice models.Invoice) error
	GetInvoiceByBookingID(bookingID string) (*models.Invoice, error)
	UpdateInvoiceStatus(bookingID string, status models.PaymentStatus) (*models.Invoice, error)
}

type paymentRepo struct{}

func NewPaymentRepo() PaymentRepo {
	return &paymentRepo{}
}

func (r *paymentRepo) CreatePayment(payment models.Payment) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("payments").
		Insert(payment, false, "", "", "").
		Execute()
	if err != nil {
		return fmt.Errorf("gagal membuat payment: %v", err)
	}
	return nil
}

func (r *paymentRepo) GetPaymentByBookingID(bookingID string) (*models.Payment, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("payments").
		Select("*", "", false).
		Eq("booking_id", bookingID).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil payment: %v", err)
	}
	var payment models.Payment
	if err := json.Unmarshal(resp, &payment); err != nil {
		return nil, fmt.Errorf("gagal decode payment: %v", err)
	}
	return &payment, nil
}

func (r *paymentRepo) UpdatePaymentStatus(bookingID string, status models.PaymentStatus, provider, reference string) (*models.Payment, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	updateData := map[string]any{
		"status": status,
	}
	if status == models.PaymentStatusPaid {
		updateData["paid_at"] = time.Now()
	}
	if provider != "" {
		updateData["provider"] = provider
	}
	if reference != "" {
		updateData["reference"] = reference
	}
	resp, _, err := config.SupabaseClient.
		From("payments").
		Update(updateData, "", "").
		Eq("booking_id", bookingID).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui payment: %v", err)
	}
	var payment models.Payment
	if err := json.Unmarshal(resp, &payment); err != nil {
		return nil, fmt.Errorf("gagal decode payment: %v", err)
	}
	return &payment, nil
}

func (r *paymentRepo) CreateInvoice(invoice models.Invoice) error {
	if config.SupabaseClient == nil {
		return fmt.Errorf("supabase client is not initialized")
	}
	_, _, err := config.SupabaseClient.
		From("invoices").
		Insert(invoice, false, "", "", "").
		Execute()
	if err != nil {
		return fmt.Errorf("gagal membuat invoice: %v", err)
	}
	return nil
}

func (r *paymentRepo) GetInvoiceByBookingID(bookingID string) (*models.Invoice, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	resp, _, err := config.SupabaseClient.
		From("invoices").
		Select("*", "", false).
		Eq("booking_id", bookingID).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil invoice: %v", err)
	}
	var invoice models.Invoice
	if err := json.Unmarshal(resp, &invoice); err != nil {
		return nil, fmt.Errorf("gagal decode invoice: %v", err)
	}
	return &invoice, nil
}

func (r *paymentRepo) UpdateInvoiceStatus(bookingID string, status models.PaymentStatus) (*models.Invoice, error) {
	if config.SupabaseClient == nil {
		return nil, fmt.Errorf("supabase client is not initialized")
	}
	updateData := map[string]any{
		"status": status,
	}
	resp, _, err := config.SupabaseClient.
		From("invoices").
		Update(updateData, "", "").
		Eq("booking_id", bookingID).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui invoice: %v", err)
	}
	var invoice models.Invoice
	if err := json.Unmarshal(resp, &invoice); err != nil {
		return nil, fmt.Errorf("gagal decode invoice: %v", err)
	}
	return &invoice, nil
}
