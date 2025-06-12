package routes

import (
	"dorm-chat-api/controllers"
	"dorm-chat-api/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {

	// ===== WebSocket без авторизации =====
	// (если нужна авторизация — заверни в middleware.Protected() аналогично HTTP)
	app.Get("/ws/:chat_id", websocket.New(func(c *websocket.Conn) {
		controllers.ChatWebSocketHandler(c, db)
	}))

	// ===== HTTP-API под JWT =====
	api := app.Group("/api", middleware.Protected())

	api.Get("/chats", controllers.GetChatsHandler(db))
	api.Get("/chats/:chat_id/messages", controllers.GetChatMessages(db))
	api.Post("/chats/:chat_id/messages", controllers.SendMessage(db))

	api.Post("/chats/init_all", controllers.CreateAllChatsHandler(db))
	api.Delete("/chats/cleanup", controllers.CleanupChatsHandler(db))
}
