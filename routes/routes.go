package routes

import (
	"dorm-chat-api/controllers"
	"dorm-chat-api/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	api := app.Group("/api", middleware.Protected())

	api.Get("/chats", controllers.GetChats(db))
	api.Get("/chats/:chat_id/messages", controllers.GetChatMessages(db))
	api.Post("/chats/:chat_id/messages", controllers.SendMessage(db))
}
