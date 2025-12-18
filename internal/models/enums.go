package models

type GuestType string

const (
	GuestTypeAdult GuestType = "Adult"
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
	BookingStatusCancel     BookingStatus = "Cancelled"
	BookingStatusCheckedIn  BookingStatus = "CheckedIn"
	BookingStatusCheckedOut BookingStatus = "CheckedOut"
	BookingStatusNoShow     BookingStatus = "NoShow"
)

type RateType string

const (
	RateTypeNonRefundableRoomOnly      RateType = "Non Refundable Room Only"
	RateTypeNonRefundableWithBreakfast RateType = "Non Refundable with Breakfast"
	RateTypeRefundableRoomOnly         RateType = "Refundable Room Only"
	RateTypeRefundableWithBreakfast    RateType = "Refundable with Breakfast"
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

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "Pending"
	PaymentStatusPaid     PaymentStatus = "Paid"
	PaymentStatusRefunded PaymentStatus = "Refunded"
)
