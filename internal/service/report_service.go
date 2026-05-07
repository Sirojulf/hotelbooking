package service

import (
	"fmt"
	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"time"
)

type ReportSummary struct {
	TotalReservations int                `json:"total_reservations"`
	Revenue           float64            `json:"revenue"`
	Occupancy         float64            `json:"occupancy"`
	ADR               float64            `json:"adr"`
	RevPAR            float64            `json:"revpar"`
	OccupancyByDate   map[string]float64 `json:"occupancy_by_date"`
}

type ReportService interface {
	GetSummary(hotelID string, start, end time.Time) (*ReportSummary, error)
}

type reportService struct {
	resRepo   repository.ReservationRepo
	hotelRepo repository.HotelRepo
}

func NewReportService(resRepo repository.ReservationRepo, hotelRepo repository.HotelRepo) ReportService {
	return &reportService{
		resRepo:   resRepo,
		hotelRepo: hotelRepo,
	}
}

func (s *reportService) GetSummary(hotelID string, start, end time.Time) (*ReportSummary, error) {
	if start.IsZero() || end.IsZero() {
		return nil, fmt.Errorf("start dan end wajib diisi")
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end tidak boleh sebelum start")
	}

	startStr := start.Format("2006-01-02")
	endStr := end.AddDate(0, 0, 1).Format("2006-01-02")

	reservations, err := s.resRepo.ListReservations(hotelID, "", startStr, endStr)
	if err != nil {
		return nil, err
	}

	rooms, err := s.hotelRepo.ListRooms(hotelID, "")
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

	occupancyByDate := make(map[string]float64)
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		occupancyByDate[day.Format("2006-01-02")] = 0
	}

	var totalNights int
	var totalReservations int
	var revenue float64

	for _, r := range reservations {
		if r.PaymentStatus == models.PaymentStatusCancelled || r.PaymentStatus == models.PaymentStatusPending {
			continue
		}
		nights := int(r.CheckOutDate.Sub(r.CheckInDate.Time).Hours() / 24)
		totalReservations++
		totalNights += nights
		revenue += r.TotalPrice

		for day := r.CheckInDate.Time; day.Before(r.CheckOutDate.Time); day = day.AddDate(0, 0, 1) {
			key := day.Format("2006-01-02")
			if _, ok := occupancyByDate[key]; ok {
				occupancyByDate[key]++
			}
		}
	}

	for k, v := range occupancyByDate {
		occupancyByDate[k] = v / float64(roomCount)
	}

	occupancy := float64(totalNights) / float64(roomCount*days)
	if occupancy < 0 {
		occupancy = 0
	}

	adr := 0.0
	if totalNights > 0 {
		adr = revenue / float64(totalNights)
	}
	revpar := revenue / float64(roomCount*days)

	return &ReportSummary{
		TotalReservations: totalReservations,
		Revenue:           revenue,
		Occupancy:         occupancy,
		ADR:               adr,
		RevPAR:            revpar,
		OccupancyByDate:   occupancyByDate,
	}, nil
}
