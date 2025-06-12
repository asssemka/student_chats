// controllers/chat_ws.go
package controllers

import (
	"log"

	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

func ChatWebSocketHandler(conn *websocket.Conn, db *gorm.DB) {
	chatID := conn.Params("chat_id") // dorm_2, floor_1_3 …

	log.Printf("WebSocket connected to chat %s", chatID) // ← теперь chatID используется

	for {
		// читаем сообщение
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break // клиент закрыл соединение
		}

		// TODO: сохранить msg в БД + расслать другим

		// эхо назад (пока)
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}
