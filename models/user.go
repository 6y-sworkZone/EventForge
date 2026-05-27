package models

import (
	"time"
)

type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Email     string     `gorm:"not null;unique" json:"email"`
	Password  string     `gorm:"not null" json:"-"`
	Name      string     `gorm:"not null" json:"name"`
	Phone     string     `json:"phone"`
	Avatar    string     `json:"avatar"`
	Role      string     `gorm:"default:user" json:"role"`
	Company   string     `json:"company"`
	Position  string     `json:"position"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
}
