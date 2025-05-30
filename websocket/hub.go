package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Chat room subscriptions
	chatRooms     map[string]map[*Client]bool
	chatRoomMutex sync.RWMutex
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		chatRooms:  make(map[string]map[*Client]bool),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected: %s", client.id)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected: %s", client.id)

				// Remove client from all chat rooms
				h.chatRoomMutex.Lock()
				for roomID, clients := range h.chatRooms {
					if _, ok := clients[client]; ok {
						delete(h.chatRooms[roomID], client)
						log.Printf("Client %s left room %s", client.id, roomID)
					}
				}
				h.chatRoomMutex.Unlock()
			}
		case message := <-h.broadcast:
			// Parse the message to determine the target room
			var msg struct {
				Type     string `json:"type"`
				RoomType string `json:"roomType"` // "dorm" or "floor"
				RoomID   string `json:"roomId"`
				Content  string `json:"content"`
			}

			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}

			// Create room key
			roomKey := msg.RoomType + ":" + msg.RoomID

			// Send message to all clients in the room
			h.chatRoomMutex.RLock()
			clients, ok := h.chatRooms[roomKey]
			h.chatRoomMutex.RUnlock()

			if !ok {
				log.Printf("Room not found: %s", roomKey)
				continue
			}

			for client := range clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(clients, client)
					h.chatRoomMutex.Lock()
					if len(clients) == 0 {
						delete(h.chatRooms, roomKey)
					} else {
						h.chatRooms[roomKey] = clients
					}
					h.chatRoomMutex.Unlock()
				}
			}
		}
	}
}

// JoinRoom adds a client to a chat room
func (h *Hub) JoinRoom(client *Client, roomType, roomID string) {
	roomKey := roomType + ":" + roomID

	h.chatRoomMutex.Lock()
	defer h.chatRoomMutex.Unlock()

	if _, ok := h.chatRooms[roomKey]; !ok {
		h.chatRooms[roomKey] = make(map[*Client]bool)
	}
	h.chatRooms[roomKey][client] = true
	log.Printf("Client %s joined room %s", client.id, roomKey)
}

// LeaveRoom removes a client from a chat room
func (h *Hub) LeaveRoom(client *Client, roomType, roomID string) {
	roomKey := roomType + ":" + roomID

	h.chatRoomMutex.Lock()
	defer h.chatRoomMutex.Unlock()

	if clients, ok := h.chatRooms[roomKey]; ok {
		delete(clients, client)
		log.Printf("Client %s left room %s", client.id, roomKey)

		if len(clients) == 0 {
			delete(h.chatRooms, roomKey)
		} else {
			h.chatRooms[roomKey] = clients
		}
	}
}
