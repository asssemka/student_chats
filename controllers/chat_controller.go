package controllers

import (
	"dorm-chat-api/models"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Получить список чатов
func GetChats(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var chats []models.Chat
		db.Find(&chats)
		return c.JSON(chats)
	}
}

// Получить сообщения по чату
func GetChatMessages(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		chatID := c.Params("chat_id")
		var messages []models.Message
		db.Where("chat_id = ?", chatID).Order("created_at asc").Find(&messages)
		return c.JSON(messages)
	}
}

// Отправить сообщение в чат
func SendMessage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		chatID := c.Params("chat_id")
		var req struct {
			Content string `json:"content"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "bad request"})
		}

		userID := c.Locals("userID")
		if userID == nil {
			return c.Status(401).JSON(fiber.Map{"error": "user_id not found in token"})
		}

		// Приводим userID к string для поля SenderID/SenderS
		senderID := fmt.Sprintf("%v", userID)

		msg := models.Message{
			ChatID:    chatID,
			SenderID:  senderID, // Если переименуешь — SenderID: senderID
			Content:   req.Content,
			CreatedAt: time.Now(),
		}
		db.Create(&msg)
		return c.JSON(msg)
	}
}
