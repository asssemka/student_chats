package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents the base user model
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	S         string    `gorm:"unique" json:"s"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Student represents a student user
type Student struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	Course    string    `json:"course"`
	RegionID  uint      `json:"region_id"`
	DormID    uint      `json:"dorm_id"`
	RoomID    uint      `json:"room_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Admin represents an admin user
type Admin struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	Role      string    `json:"role"`
	DormID    uint      `json:"dorm_id"` // Which dormitory this admin manages
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeSave hashes the password before saving
func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return errors.New("invalid password")
	}
	return nil
}
