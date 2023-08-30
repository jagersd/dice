package main

import "log"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

type Message struct {
	table   string
	content []byte
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			activeTables[client.table].wsConnections[client] = true
		case client := <-h.unregister:
			log.Println("player disconnected")
			client.send <- []byte("A player has disconnected")
			if _, ok := activeTables[client.table].wsConnections[client]; ok {
				delete(activeTables[client.table].wsConnections, client)
				close(client.send)
			}
			if len(activeTables[client.table].wsConnections) == 0 {
				log.Println("Removing table ", client.table)
				delete(activeTables, client.table)
			}
		case message := <-h.broadcast:
			for client := range activeTables[message.table].wsConnections {
				select {
				case client.send <- message.content:
				default:
					close(client.send)
					delete(activeTables[message.table].wsConnections, client)
				}
			}
		}
	}
}
