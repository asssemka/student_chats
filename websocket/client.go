package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Unique client ID
	id string

	// User ID from authentication
	userID uint
}

// Message represents a chat message
type Message struct {
	Type     string    `json:"type"`
	RoomType string    `json:"roomType"`
	RoomID   string    `json:"roomId"`
	UserID   uint      `json:"userId"`
	Content  string    `json:"content"`
	Time     time.Time `json:"time"`
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Parse the message
		var msg struct {
			Type     string `json:"type"`
			RoomType string `json:"roomType"`
			RoomID   string `json:"roomId"`
			Content  string `json:"content,omitempty"`
		}

		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "join":
			c.hub.JoinRoom(c, msg.RoomType, msg.RoomID)
		case "leave":
			c.hub.LeaveRoom(c, msg.RoomType, msg.RoomID)
		case "message":
			// Create a new message with user ID and timestamp
			newMsg := Message{
				Type:     "message",
				RoomType: msg.RoomType,
				RoomID:   msg.RoomID,
				UserID:   c.userID,
				Content:  msg.Content,
				Time:     time.Now(),
			}

			// Marshal the message
			msgBytes, err := json.Marshal(newMsg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			// Broadcast the message
			c.hub.broadcast <- msgBytes
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer
func ServeWs(hub *Hub, conn *websocket.Conn, userID uint) {
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		id:     uuid.New().String(),
		userID: userID,
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines
	go client.writePump()
	go client.readPump()
}
