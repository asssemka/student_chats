package controllers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"

	ws "dorm-chat-api/websocket"
)

// WebsocketController handles websocket connections
type WebsocketController struct {
	Hub *ws.Hub
}

// NewWebsocketController creates a new websocket controller
func NewWebsocketController(hub *ws.Hub) *WebsocketController {
	return &WebsocketController{Hub: hub}
}

// HandleWebsocket handles websocket connections
func (c *WebsocketController) HandleWebsocket(ctx *fiber.Ctx) error {
	// Get user from token
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["id"].(float64))

	// Upgrade to websocket
	return websocket.New(func(conn *websocket.Conn) {
		// Handle websocket connection
		ws.ServeWs(c.Hub, conn, userID)
	})(ctx)
}
