package ws

import (
	"log"
	"sync"
)

type Hub struct {
	clients map[*Client]bool

	broadcast chan []byte

	register chan *Client

	unregister chan *Client

	monitors map[string]map[*Client]bool

	monitorMutex sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		monitors:   make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if client.Role == "TEACHER" || client.Role == "ADMIN" {
				h.registerMonitor(client)
			}
			log.Printf("Client connected: %s (Role: %s)", client.UserID, client.Role)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if client.Role == "TEACHER" || client.Role == "ADMIN" {
					h.unregisterMonitor(client)
				}
				log.Printf("Client disconnected: %s", client.UserID)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) registerMonitor(client *Client) {
	h.monitorMutex.Lock()
	defer h.monitorMutex.Unlock()

	if _, ok := h.monitors["monitor"]; !ok {
		h.monitors["monitor"] = make(map[*Client]bool)
	}
	h.monitors["monitor"][client] = true
}

func (h *Hub) unregisterMonitor(client *Client) {
	h.monitorMutex.Lock()
	defer h.monitorMutex.Unlock()

	if group, ok := h.monitors["monitor"]; ok {
		delete(group, client)
	}
}

func (h *Hub) BroadcastToMonitors(message []byte) {
	h.monitorMutex.RLock()
	defer h.monitorMutex.RUnlock()

	if group, ok := h.monitors["monitor"]; ok {
		for client := range group {
			select {
			case client.send <- message:
			default:
			}
		}
	}
}
