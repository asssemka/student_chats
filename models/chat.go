package models

import "time"

type Chat struct {
	ID        uint   `gorm:"primaryKey"`
	ChatID    string `gorm:"uniqueIndex"` // Например: dorm_1, floor_2_3
	Type      string // dorm или floor
	DormID    uint
	Floor     uint
	Name      string
	CreatedAt time.Time
}
