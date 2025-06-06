package controllers

import (
	ws "dorm-chat-api/websocket"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func ChatWebSocket(hub *ws.Hub) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		ws.ServeWS(hub, c)
	})
}
