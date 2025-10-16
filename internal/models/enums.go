package models

type GuestType string

const (
	GuestTypedult  GuestType = "Adult"
	GuestTypeChild GuestType = "Child"
)

type Gender string

const (
	GenderMale   Gender = "Male"
	GenderFemale Gender = "Female"
)

type VIPStatus string

const (
	VIPStatusBronze   VIPStatus = "Bronze"
	VIPStatusSilver   VIPStatus = "Silver"
	VIPStatusGold     VIPStatus = "Gold"
	VIPStatusPlatinum VIPStatus = "Platinum"
)

type BookingStatus string

const (
	BookingStatusNew        BookingStatus = "New"
	BookingStatusConfirmed  BookingStatus = "Confirmed"
	BookingStatusCancel     BookingStatus = "Cancel"
	BookingStatusCheckedIn  BookingStatus = "CheckedIn"
	BookingStatusCheckedOut BookingStatus = "CheckedOut"
	BookingStatusNoShow     BookingStatus = "NoShow"
)

type RoomType string

const (
	RoomTypeStandard RoomType = "Standard"
	RoomTypeDeluxe   RoomType = "Deluxe"
	RoomTypeSuite    RoomType = "Suite"
)

type RoomStatus string

const (
	RoomStatusAvailable  RoomStatus = "Available"
	RoomStatusOccupied   RoomStatus = "Occupied"
	RoomStatusOutOfOrder RoomStatus = "OutOfOrder"
)

type HousekeepingStatus string

const (
	HousekeepingStatusClean        HousekeepingStatus = "Clean"
	HousekeepingStatusDirty        HousekeepingStatus = "Dirty"
	HousekeepingStatusInspected    HousekeepingStatus = "Inspected"
	HousekeepingStatusPickup       HousekeepingStatus = "Pickup"
	HousekeepingStatusOutOfOrder   HousekeepingStatus = "OutOfOrder"
	HousekeepingStatusOutOfService HousekeepingStatus = "OutOfService"
)
