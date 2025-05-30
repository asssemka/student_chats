package models

import (
	"time"
)

// DormChat represents a chat for an entire dormitory
type DormChat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DormID    uint      `json:"dorm_id"`
	Dorm      Dorm      `gorm:"foreignKey:DormID" json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FloorChat represents a chat for a specific floor in a dormitory
type FloorChat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DormID    uint      `json:"dorm_id"`
	Dorm      Dorm      `gorm:"foreignKey:DormID" json:"-"`
	Floor     int       `json:"floor"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a chat message
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Content    string    `json:"content"`
	UserID     uint      `json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"user"`
	DormChatID *uint     `json:"dorm_chat_id,omitempty"`
	DormChat   *DormChat `gorm:"foreignKey:DormChatID" json:"-"`
	FloorChatID *uint     `json:"floor_chat_id,omitempty"`
	FloorChat  *FloorChat `gorm:"foreignKey:FloorChatID" json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

// ChatMember represents a user's membership in a chat
type ChatMember struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"-"`
	DormChatID *uint     `json:"dorm_chat_id,omitempty"`
	DormChat   *DormChat `gorm:"foreignKey:DormChatID" json:"-"`
	FloorChatID *uint     `json:"floor_chat_id,omitempty"`
	FloorChat  *FloorChat `gorm:"foreignKey:FloorChatID" json:"-"`
	IsAdmin    bool      `json:"is_admin"`
	CreatedAt  time.Time `json:"created_at"`
}
