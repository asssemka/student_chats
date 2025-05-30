package models

import (
	"time"
)

// Dorm represents a dormitory
type Dorm struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	TotalPlaces int       `json:"total_places"`
	Cost        int       `json:"cost"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Room represents a room in a dormitory
type Room struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DormID    uint      `json:"dorm_id"`
	Dorm      Dorm      `gorm:"foreignKey:DormID" json:"-"`
	Number    string    `json:"number"`
	Capacity  int       `json:"capacity"`
	Floor     int       `json:"floor"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
