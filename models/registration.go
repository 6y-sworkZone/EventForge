package models

import (
	"time"
)

type RegistrationStatus string

const (
	RegistrationStatusPending   RegistrationStatus = "pending"
	RegistrationStatusConfirmed RegistrationStatus = "confirmed"
	RegistrationStatusWaitlist  RegistrationStatus = "waitlist"
	RegistrationStatusCancelled RegistrationStatus = "cancelled"
	RegistrationStatusCheckedIn RegistrationStatus = "checked_in"
)

type Registration struct {
	ID              uint               `gorm:"primaryKey" json:"id"`
	EventID         uint               `gorm:"not null" json:"event_id"`
	TicketID        *uint              `json:"ticket_id"`
	UserID          uint               `json:"user_id"`
	Name            string             `gorm:"not null" json:"name"`
	Email           string             `gorm:"not null" json:"email"`
	Phone           string             `json:"phone"`
	Company         string             `json:"company"`
	Position        string             `json:"position"`
	DietPreference  string             `json:"diet_preference"`
	CustomFields    string             `json:"custom_fields"`
	Status          RegistrationStatus `gorm:"default:pending" json:"status"`
	QRCode          string             `json:"qr_code"`
	CheckedInAt     *time.Time         `json:"checked_in_at"`
	WaitlistOrder   int                `gorm:"default:0" json:"waitlist_order"`
	OrderID         *uint              `json:"order_id"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	DeletedAt       *time.Time         `gorm:"index" json:"-"`

	Event  Event   `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Ticket *Ticket `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	Order  *Order  `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}
