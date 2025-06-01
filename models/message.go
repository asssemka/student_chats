package models

import "time"

type Message struct {
	ID        uint   `gorm:"primaryKey"`
	ChatID    string // Соответствует Chat.ChatID
	SenderID  string // s из Django (user-идентификатор)
	Content   string
	CreatedAt time.Time
}
