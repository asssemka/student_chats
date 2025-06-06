package models

import "time"

type Chat struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    ChatID    string    `gorm:"uniqueIndex" json:"chatID"`
    Type      string    `json:"type"`
    DormID    uint      `json:"dormID"`
    Floor     uint      `json:"floor"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"createdAt"`
}
