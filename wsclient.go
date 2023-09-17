package main

import (
	"dice/html"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub    *Hub
	table  string
	conn   *websocket.Conn
	send   chan []byte
	player *player
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var message map[string]interface{}
		err := c.conn.ReadJSON(&message)
		if err != nil {
			log.Printf("error: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.hub.broadcast <- c.generateResponse(message)
	}
}

func (c *Client) generateResponse(incomingMessage map[string]interface{}) Response {
	for k := range incomingMessage {
		switch k {
		case "start-roll":
			c.player.roll(true)
			c.send <- []byte(`<div id="player-control">Wait for shooter to be determined</div>`)
			// c.send <- html.ShowWagerControlls(c.player)
			activeTables[c.table].determineShooter()
		case "bet":
			if _, ok := incomingMessage["betfor"]; ok {
				c.player.setWager(parseInterface(incomingMessage["betfor"], "s").s)
				c.player.placeBet(parseInterface(incomingMessage["wager"], "i").i)
			}
			if c.player.IsShooter {
				activeTables[c.table].BetHight = uint(parseInterface(incomingMessage["wager"], "i").i)
				activeTables[c.table].letNonShootersBet()
			}
			c.send <- html.Play(c.player)
		case "shooter-roll":
			c.player.roll(false)
			activeTables[c.table].evaluateRoll()
		}
	}

	var response Response

	response.table = c.table
	response.update = true
	response.content = html.WSGameState(activeTables[c.table])

	return response
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
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
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
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

func serveWs(t *table, index int, hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:    hub,
		table:  t.InternalName,
		conn:   conn,
		send:   make(chan []byte, 256),
		player: &t.Players[index],
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
