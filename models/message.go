package models

import "time"

type Message struct {
    ID         uint      `gorm:"primaryKey"`
    ChatID     string    // Соответствует Chat.ChatID
    SenderID   string    // user-id из Django (может быть либо студента, либо администратора)
    SenderType string    // "student" или "admin" (добавили)
    Content    string
    CreatedAt  time.Time
}