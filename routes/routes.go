package routes

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"dorm-chat-api/controllers"
	"dorm-chat-api/middleware"
	ws "dorm-chat-api/websocket"
)

// SetupRoutes sets up all the routes for the application
func SetupRoutes(app *fiber.App, db *gorm.DB) {
	// Create websocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Create controllers
	authController := controllers.NewAuthController(db)
	chatController := controllers.NewChatController(db)

	// Auth routes
	auth := app.Group("/api/auth")
	auth.Post("/login", authController.Login)

	// Protected routes
	api := app.Group("/api", middleware.Protected())

	// Chat routes
	chats := api.Group("/chats")
	chats.Get("/", chatController.GetAvailableChats)
	chats.Get("/:type/:id", chatController.GetChatMessages)
	chats.Post("/:type/:id", chatController.SendMessage)

	// Websocket route with authentication middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		// Check if it's a websocket upgrade request
		if websocket.IsWebSocketUpgrade(c) {
			return middleware.Protected()(c)
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(conn *websocket.Conn) {
		// Get user from context (set by middleware)
		user := conn.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userID := uint(claims["id"].(float64))

		ws.ServeWs(hub, conn, userID)
	}))
}
