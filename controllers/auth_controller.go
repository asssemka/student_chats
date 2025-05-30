package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"dorm-chat-api/models"
	"dorm-chat-api/utils"
)

// AuthController handles authentication related requests
type AuthController struct {
	DB *gorm.DB
}

// LoginRequest represents the login request body
type LoginRequest struct {
	S        string `json:"s"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// NewAuthController creates a new auth controller
func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

// Login handles user login
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var req LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Find user by S identifier
	var user models.User
	if err := c.DB.Where("s = ?", req.S).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Check password
	if err := user.CheckPassword(req.Password); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Check if user is a student
	var student models.Student
	isStudent := c.DB.Where("user_id = ?", user.ID).First(&student).Error == nil

	// Check if user is an admin
	var admin models.Admin
	isAdmin := c.DB.Where("user_id = ?", user.ID).First(&admin).Error == nil

	// Generate JWT token
	claims := jwt.MapClaims{
		"id":        user.ID,
		"s":         user.S,
		"isStudent": isStudent,
		"isAdmin":   isAdmin,
		"exp":       time.Now().Add(time.Hour * 72).Unix(),
	}

	if isStudent {
		claims["student_id"] = student.UserID
		claims["dorm_id"] = student.DormID
		claims["room_id"] = student.RoomID
	}

	if isAdmin {
		claims["admin_id"] = admin.UserID
		claims["role"] = admin.Role
		claims["dorm_id"] = admin.DormID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(utils.GetEnv("JWT_SECRET", "secret")))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not generate token")
	}

	// Return response
	var userData interface{}
	if isStudent {
		c.DB.Preload("User").First(&student, "user_id = ?", user.ID)
		userData = student
	} else if isAdmin {
		c.DB.Preload("User").First(&admin, "user_id = ?", user.ID)
		userData = admin
	} else {
		userData = user
	}

	return ctx.JSON(LoginResponse{
		Token: t,
		User:  userData,
	})
}
