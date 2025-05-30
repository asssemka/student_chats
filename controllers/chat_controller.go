package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"dorm-chat-api/models"
)

// ChatController handles chat related requests
type ChatController struct {
	DB *gorm.DB
}

// NewChatController creates a new chat controller
func NewChatController(db *gorm.DB) *ChatController {
	return &ChatController{DB: db}
}

// GetAvailableChats returns all chats available to the current user
func (c *ChatController) GetAvailableChats(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["id"].(float64))
	isStudent := claims["isStudent"].(bool)
	isAdmin := claims["isAdmin"].(bool)

	type ChatResponse struct {
		ID        uint      `json:"id"`
		Type      string    `json:"type"` // "dorm" or "floor"
		Name      string    `json:"name"`
		DormID    uint      `json:"dorm_id"`
		Floor     *int      `json:"floor,omitempty"`
		CreatedAt time.Time `json:"created_at"`
	}

	var chats []ChatResponse

	if isStudent {
		// Get student's dorm and floor
		var student models.Student
		if err := c.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "Student not found")
		}

		// Get student's room to determine floor
		var room models.Room
		if err := c.DB.Where("id = ?", student.RoomID).First(&room).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "Room not found")
		}

		// Get dorm chat
		var dormChat models.DormChat
		if err := c.DB.Where("dorm_id = ?", student.DormID).First(&dormChat).Error; err == nil {
			chats = append(chats, ChatResponse{
				ID:        dormChat.ID,
				Type:      "dorm",
				Name:      dormChat.Name,
				DormID:    dormChat.DormID,
				CreatedAt: dormChat.CreatedAt,
			})
		}

		// Get floor chat
		var floorChat models.FloorChat
		if err := c.DB.Where("dorm_id = ? AND floor = ?", student.DormID, room.Floor).First(&floorChat).Error; err == nil {
			floor := room.Floor
			chats = append(chats, ChatResponse{
				ID:        floorChat.ID,
				Type:      "floor",
				Name:      floorChat.Name,
				DormID:    floorChat.DormID,
				Floor:     &floor,
				CreatedAt: floorChat.CreatedAt,
			})
		}
	} else if isAdmin {
		// Get admin's dorm
		var admin models.Admin
		if err := c.DB.Where("user_id = ?", userID).First(&admin).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "Admin not found")
		}

		// Get all dorm chats for admin's dorm
		var dormChat models.DormChat
		if err := c.DB.Where("dorm_id = ?", admin.DormID).First(&dormChat).Error; err == nil {
			chats = append(chats, ChatResponse{
				ID:        dormChat.ID,
				Type:      "dorm",
				Name:      dormChat.Name,
				DormID:    dormChat.DormID,
				CreatedAt: dormChat.CreatedAt,
			})
		}

		// Get all floor chats for admin's dorm
		var floorChats []models.FloorChat
		if err := c.DB.Where("dorm_id = ?", admin.DormID).Find(&floorChats).Error; err == nil {
			for _, fc := range floorChats {
				floor := fc.Floor
				chats = append(chats, ChatResponse{
					ID:        fc.ID,
					Type:      "floor",
					Name:      fc.Name,
					DormID:    fc.DormID,
					Floor:     &floor,
					CreatedAt: fc.CreatedAt,
				})
			}
		}
	}

	return ctx.JSON(fiber.Map{
		"chats": chats,
	})
}

// GetChatMessages returns all messages for a specific chat
func (c *ChatController) GetChatMessages(ctx *fiber.Ctx) error {
	chatType := ctx.Params("type") // "dorm" or "floor"
	chatID, err := ctx.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid chat ID")
	}

	// Verify user has access to this chat
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["id"].(float64))
	isStudent := claims["isStudent"].(bool)
	isAdmin := claims["isAdmin"].(bool)

	if isStudent {
		var student models.Student
		if err := c.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "Student not found")
		}

		// Get student's room to determine floor
		var room models.Room
		if err := c.DB.Where("id = ?", student.RoomID).First(&room).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "Room not found")
		}

		// Check if student has access to this chat
		if chatType == "dorm" {
			var dormChat models.DormChat
			if err := c.DB.Where("id = ? AND dorm_id = ?", chatID, student.DormID).First(&dormChat).Error; err != nil {
				return fiber.NewError(fiber.StatusForbidden, "Access denied")
			}
		} else if chatType == "floor" {
			var floorChat models.FloorChat
			if err := c.DB.Where("id = ? AND dorm_id = ? AND floor = ?", chatID, student.DormID, room.Floor).First(&floorChat).Error; err != nil {
				return fiber.NewError(fiber.StatusForbidden, "Access denied")
			}
		} else {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid chat type")
		}
	} else if isAdmin {
		var admin models.Admin
		if err := c.DB.Where("user_id = ?", userID).First(&admin).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "Admin not found")
		}

		// Check if admin has access to this chat
		if chatType == "dorm" {
			var dormChat models.DormChat
			if err := c.DB.Where("id = ? AND dorm_id = ?", chatID, admin.DormID).First(&dormChat).Error; err != nil {
				return fiber.NewError(fiber.StatusForbidden, "Access denied")
			}
		} else if chatType == "floor" {
			var floorChat models.FloorChat
			if err := c.DB.Where("id = ? AND dorm_id = ?", chatID, admin.DormID).First(&floorChat).Error; err != nil {
				return fiber.NewError(fiber.StatusForbidden, "Access denied")
			}
		} else {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid chat type")
		}
	}

	// Get messages
	var messages []models.Message
	query := c.DB.Preload("User")

	if chatType == "dorm" {
		query = query.Where("dorm_chat_id = ?", chatID)
	} else if chatType == "floor" {
		query = query.Where("floor_chat_id = ?", chatID)
	}

	if err := query.Order("created_at DESC").Limit(100).Find(&messages).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch messages")
	}

	return ctx.JSON(fiber.Map{
		"messages": messages,
	})
}

// SendMessage sends a new message to a chat
func (c *ChatController) SendMessage(ctx *fiber.Ctx) error {
	chatType := ctx.Params("type") // "dorm" or "floor"
	chatID, err := ctx.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid chat ID")
	}

	// Get user from token
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["id"].(float64))

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Create message
	message := models.Message{
		Content: req.Content,
		UserID:  userID,
	}

	if chatType == "dorm" {
		id := uint(chatID)
		message.DormChatID = &id
	} else if chatType == "floor" {
		id := uint(chatID)
		message.FloorChatID = &id
	} else {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid chat type")
	}

	if err := c.DB.Create(&message).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send message")
	}

	// Load user data for response
	c.DB.Preload("User").First(&message, message.ID)

	return ctx.JSON(message)
}
