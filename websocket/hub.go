package websocket

import (
	"log"
)

type Hub struct {
	// Карта: chatID -> set of clients
	Rooms map[string]map[*Client]bool

	// Отправка сообщения в комнату
	Broadcast chan BroadcastMessage

	// Регистрация клиента
	Register chan *Client

	// Отключение клиента
	Unregister chan *Client
}

type BroadcastMessage struct {
	ChatID  string
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Broadcast:  make(chan BroadcastMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if h.Rooms[client.ChatID] == nil {
				h.Rooms[client.ChatID] = make(map[*Client]bool)
			}
			h.Rooms[client.ChatID][client] = true
			log.Printf("WS client joined: chat=%s user=%s", client.ChatID, client.UserS)
		case client := <-h.Unregister:
			if clients, ok := h.Rooms[client.ChatID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
				}
				if len(clients) == 0 {
					delete(h.Rooms, client.ChatID)
				}
			}
		case message := <-h.Broadcast:
			if clients, ok := h.Rooms[message.ChatID]; ok {
				for client := range clients {
					select {
					case client.Send <- message.Message:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
		}
	}
}
