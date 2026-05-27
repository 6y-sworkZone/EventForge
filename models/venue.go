package models

import (
	"time"
)

type Venue struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"not null" json:"name"`
	Address      string     `json:"address"`
	Capacity     int        `gorm:"default:0" json:"capacity"`
	Facilities   string     `json:"facilities"`
	FloorPlan    string     `json:"floor_plan"`
	Description  string     `json:"description"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`

	Rooms []VenueRoom `gorm:"foreignKey:VenueID" json:"rooms,omitempty"`
}

type VenueRoom struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	VenueID   uint       `gorm:"not null" json:"venue_id"`
	Name      string     `gorm:"not null" json:"name"`
	Capacity  int        `gorm:"default:0" json:"capacity"`
	Floor     string     `json:"floor"`
	SeatMap   string     `json:"seat_map"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	Venue Venue `gorm:"foreignKey:VenueID" json:"venue,omitempty"`
}

type Seat struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	VenueRoomID uint      `gorm:"not null" json:"venue_room_id"`
	SeatNumber  string    `gorm:"not null" json:"seat_number"`
	Row         string    `json:"row"`
	Section     string    `json:"section"`
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	Status      string    `gorm:"default:available" json:"status"`
	EventID     *uint     `json:"event_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	VenueRoom VenueRoom `gorm:"foreignKey:VenueRoomID" json:"venue_room,omitempty"`
}
