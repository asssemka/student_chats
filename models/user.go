package models

import (
	"time"
)

// UserRole типы пользователей
type UserRole string

const (
	RoleStudent UserRole = "student"
	RoleAdmin   UserRole = "admin"
)

type User struct {
	ID        uint     `gorm:"primaryKey"`
	S         string   `gorm:"unique;not null"` // Совпадает с полем "s" из Django
	Role      UserRole `gorm:"not null"`
	CreatedAt time.Time
}

// Таблица пользователей нужна только для чатов и авторизации.
// Если синхронизация с Django по "s" (логину), можно автоматически создавать запись при первом входе.
