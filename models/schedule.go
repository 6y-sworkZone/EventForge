package models

import (
	"time"
)

type Schedule struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	EventID   uint       `gorm:"not null" json:"event_id"`
	Name      string     `gorm:"not null" json:"name"`
	Date      time.Time  `gorm:"not null" json:"date"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	Event      Event      `gorm:"foreignKey:EventID" json:"event,omitempty"`
	AgendaItems []AgendaItem `gorm:"foreignKey:ScheduleID" json:"agenda_items,omitempty"`
}

type AgendaItem struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ScheduleID  uint       `gorm:"not null" json:"schedule_id"`
	SpeakerID   *uint      `json:"speaker_id"`
	VenueRoomID *uint      `json:"venue_room_id"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `json:"description"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	Location    string     `json:"location"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"-"`

	Schedule   Schedule   `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	Speaker    *Speaker   `gorm:"foreignKey:SpeakerID" json:"speaker,omitempty"`
	VenueRoom  *VenueRoom `gorm:"foreignKey:VenueRoomID" json:"venue_room,omitempty"`
}

type Speaker struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	EventID   uint       `gorm:"not null" json:"event_id"`
	Name      string     `gorm:"not null" json:"name"`
	Avatar    string     `json:"avatar"`
	Title     string     `json:"title"`
	Company   string     `json:"company"`
	Bio       string     `json:"bio"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	Event Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
}
