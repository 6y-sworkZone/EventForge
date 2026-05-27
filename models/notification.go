package models

import (
	"time"
)

type NotificationType string

const (
	NotificationTypeRegistrationSuccess NotificationType = "registration_success"
	NotificationTypeEventUpdate         NotificationType = "event_update"
	NotificationTypeEventReminder       NotificationType = "event_reminder"
	NotificationTypeCheckIn             NotificationType = "check_in"
	NotificationTypeFeedback            NotificationType = "feedback"
	NotificationTypeCustom              NotificationType = "custom"
)

type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
)

type Notification struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	EventID   *uint            `json:"event_id"`
	UserID    *uint            `json:"user_id"`
	Type      NotificationType `gorm:"not null" json:"type"`
	Subject   string           `gorm:"not null" json:"subject"`
	Content   string           `gorm:"not null" json:"content"`
	Status    NotificationStatus `gorm:"default:pending" json:"status"`
	ScheduledAt *time.Time     `json:"scheduled_at"`
	SentAt    *time.Time       `json:"sent_at"`
	ErrorMsg  string           `json:"error_msg"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`

	Event *Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
}

type NotificationTemplate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;unique" json:"name"`
	Type      string    `json:"type"`
	Subject   string    `gorm:"not null" json:"subject"`
	Content   string    `gorm:"not null" json:"content"`
	Variables string    `json:"variables"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
