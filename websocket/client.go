package websocket

import (
	"log"

	"github.com/gofiber/contrib/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	Hub    *Hub
	ChatID string
	UserS  string
	Send   chan []byte
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("WS read error:", err)
			break
		}
		// Простая ретрансляция в комнату
		c.Hub.Broadcast <- BroadcastMessage{
			ChatID:  c.ChatID,
			Message: msg,
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				// Канал закрыт
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

// Вспомогательная функция для старта клиента
func ServeWS(hub *Hub, conn *websocket.Conn) {
	// Получаем chatID и userS из query params
	chatID := conn.Params("chat_id")
	userS := conn.Query("userS")

	client := &Client{
		Conn:   conn,
		Hub:    hub,
		ChatID: chatID,
		UserS:  userS,
		Send:   make(chan []byte, 256),
	}
	client.Hub.Register <- client

	go client.WritePump()
	client.ReadPump()
}
