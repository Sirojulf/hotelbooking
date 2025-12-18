package service

import (
	"fmt"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"time"
)

type ReportSummary struct {
	TotalBookings int     `json:"total_bookings"`
	Revenue       float64 `json:"revenue"`
	Occupancy     float64 `json:"occupancy"`
	ADR           float64 `json:"adr"`
	RevPAR        float64 `json:"revpar"`
	OccupancyByDate map[string]float64 `json:"occupancy_by_date"`
}

type ReportService interface {
	GetSummary(propertyID string, start, end time.Time) (*ReportSummary, error)
}

type reportService struct {
	bookingRepo repository.BookingRepo
	propRepo    repository.PropertyRepo
}

func NewReportService(bookingRepo repository.BookingRepo, propRepo repository.PropertyRepo) ReportService {
	return &reportService{
		bookingRepo: bookingRepo,
		propRepo:    propRepo,
	}
}

func (s *reportService) GetSummary(propertyID string, start, end time.Time) (*ReportSummary, error) {
	if start.IsZero() || end.IsZero() {
		return nil, fmt.Errorf("start dan end wajib diisi")
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end tidak boleh sebelum start")
	}
	var startStr, endStr string
	if !start.IsZero() {
		startStr = start.Format("2006-01-02")
	}
	if !end.IsZero() {
		endStr = end.Format("2006-01-02")
	}

	bookings, err := s.bookingRepo.ListBookings(propertyID, "", startStr, endStr)
	if err != nil {
		return nil, err
	}

	rooms, err := s.propRepo.ListRooms(propertyID, "")
	if err != nil {
		return nil, err
	}

	days := int(end.Sub(start).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}
	roomCount := len(rooms)
	if roomCount == 0 {
		roomCount = 1
	}

	var totalNights int
	var revenue float64
	occupancyByDate := make(map[string]float64)
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		occupancyByDate[day.Format("2006-01-02")] = 0
	}

	for _, b := range bookings {
		if b.Status == models.BookingStatusCancel || b.Status == models.BookingStatusNew {
			continue
		}
		totalNights += b.Nights
		revenue += b.TotalPrice - b.RefundAmount
		for day := b.CheckIn; day.Before(b.CheckOut); day = day.AddDate(0, 0, 1) {
			key := day.Format("2006-01-02")
			if _, ok := occupancyByDate[key]; ok {
				occupancyByDate[key] += 1
			}
		}
	}

	occupancy := float64(totalNights) / float64(roomCount*days)
	if occupancy < 0 {
		occupancy = 0
	}
	for k, v := range occupancyByDate {
		occupancyByDate[k] = v / float64(roomCount)
	}

	adr := 0.0
	if totalNights > 0 {
		adr = revenue / float64(totalNights)
	}
	revpar := revenue / float64(roomCount*days)

	return &ReportSummary{
		TotalBookings: len(bookings),
		Revenue:       revenue,
		Occupancy:     occupancy,
		ADR:           adr,
		RevPAR:        revpar,
		OccupancyByDate: occupancyByDate,
	}, nil
}
