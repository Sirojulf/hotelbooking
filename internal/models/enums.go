package models

type RoomStatus string

const (
	RoomStatusAvailable  RoomStatus = "available"
	RoomStatusOccupied   RoomStatus = "occupied"
	RoomStatusMaintenance RoomStatus = "maintenance"
)

type CleanStatus string

const (
	CleanStatusClean        CleanStatus = "clean"
	CleanStatusDirty        CleanStatus = "dirty"
	CleanStatusInspected    CleanStatus = "inspected"
	CleanStatusPickup       CleanStatus = "pickup"
	CleanStatusOutOfOrder   CleanStatus = "out_of_order"
	CleanStatusOutOfService CleanStatus = "out_of_service"
)

type GuestTier string

const (
	GuestTierBronze   GuestTier = "bronze"
	GuestTierSilver   GuestTier = "silver"
	GuestTierGold     GuestTier = "gold"
	GuestTierPlatinum GuestTier = "platinum"
)

type BookingSource string

const (
	BookingSourceWalkIn    BookingSource = "walk_in"
	BookingSourceOnline    BookingSource = "online"
	BookingSourcePhone     BookingSource = "phone"
	BookingSourceOTA       BookingSource = "ota"
	BookingSourceCorporate BookingSource = "corporate"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

type TransactionType string

const (
	TransactionTypeCharge  TransactionType = "charge"
	TransactionTypePayment TransactionType = "payment"
	TransactionTypeRefund  TransactionType = "refund"
)

type HotelStatus string

const (
	HotelStatusActive      HotelStatus = "active"
	HotelStatusMaintenance HotelStatus = "maintenance"
	HotelStatusSuspended   HotelStatus = "suspended"
)
