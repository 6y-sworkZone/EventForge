package models

import (
	"time"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusOpen      EventStatus = "open"
	EventStatusOngoing   EventStatus = "ongoing"
	EventStatusEnded     EventStatus = "ended"
	EventStatusCancelled EventStatus = "cancelled"
)

type EventType string

const (
	EventTypeMeeting   EventType = "meeting"
	EventTypeParty     EventType = "party"
	EventTypeWorkshop  EventType = "workshop"
	EventTypeExpo      EventType = "expo"
	EventTypeRace      EventType = "race"
	EventTypeOther     EventType = "other"
)

type Event struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	Title        string      `gorm:"not null" json:"title"`
	Description  string      `json:"description"`
	CoverImage   string      `json:"cover_image"`
	Type         EventType   `gorm:"not null" json:"type"`
	StartTime    time.Time   `gorm:"not null" json:"start_time"`
	EndTime      time.Time   `gorm:"not null" json:"end_time"`
	Location     string      `json:"location"`
	VenueID      *uint       `json:"venue_id"`
	Venue        *Venue      `gorm:"foreignKey:VenueID" json:"venue,omitempty"`
	MaxCapacity  int         `gorm:"default:0" json:"max_capacity"`
	IsPublic     bool        `gorm:"default:true" json:"is_public"`
	IsPaid       bool        `gorm:"default:false" json:"is_paid"`
	Status       EventStatus `gorm:"default:draft" json:"status"`
	CreatedBy    uint        `json:"created_by"`
	TemplateID   *uint       `json:"template_id"`
	IsTemplate   bool        `gorm:"default:false" json:"is_template"`
	TemplateName string      `json:"template_name"`
	ClonedFromID *uint       `json:"cloned_from_id"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	DeletedAt    *time.Time  `gorm:"index" json:"-"`

	Registrations []Registration `gorm:"foreignKey:EventID" json:"registrations,omitempty"`
	Tickets       []Ticket       `gorm:"foreignKey:EventID" json:"tickets,omitempty"`
	Schedules     []Schedule     `gorm:"foreignKey:EventID" json:"schedules,omitempty"`
	Speakers      []Speaker      `gorm:"foreignKey:EventID" json:"speakers,omitempty"`
	Notifications []Notification `gorm:"foreignKey:EventID" json:"notifications,omitempty"`
}

type CustomField struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EventID   uint      `gorm:"not null" json:"event_id"`
	Label     string    `gorm:"not null" json:"label"`
	FieldType string    `gorm:"not null;default:text" json:"field_type"`
	Required  bool      `gorm:"default:false" json:"required"`
	Options   string    `json:"options"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Feedback struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EventID   uint      `gorm:"not null" json:"event_id"`
	UserID    uint      `json:"user_id"`
	Rating    int       `gorm:"not null" json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`

	Event Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
}
