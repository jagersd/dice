package main

import "log"

type Hub struct {
	broadcast  chan Response
	register   chan *Client
	unregister chan *Client
}

type Response struct {
	table   string
	update  bool
	content []byte
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Response),
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
		case response := <-h.broadcast:
			if response.update {
				activeTables[response.table].broadcastGameState()
			} else {
				for client := range activeTables[response.table].wsConnections {
					select {
					case client.send <- response.content:
					default:
						close(client.send)
						delete(activeTables[response.table].wsConnections, client)
					}
				}
			}
		}
	}
}
