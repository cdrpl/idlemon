package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	wsHub *WsHub
	conn  *websocket.Conn
	send  chan []byte // Buffered channel of outbound messages.
}

func (c *WsClient) readPump() {
	defer func() {
		c.wsHub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(WS_MAX_MESSAGE_SIZE)
	c.conn.SetReadDeadline(time.Now().Add(WS_PONG_TIMEOUT))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(WS_PONG_TIMEOUT)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, []byte("\n"), []byte(" "), -1))

		c.wsHub.broadcast <- message
	}
}

func (c *WsClient) writePump() {
	ticker := time.NewTicker(WS_PING_PERIOD)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(WS_WRITE_TIMOUT))
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
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WS_WRITE_TIMOUT))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func WsConnectHandler(wsHub *WsHub, w http.ResponseWriter, r *http.Request) {
	conn, err := wsHub.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &WsClient{wsHub: wsHub, conn: conn, send: make(chan []byte, 256)}
	client.wsHub.register <- client

	go client.writePump()
	go client.readPump()

	log.Println("websocket client connected")
}

func CreateWsHub(upgrader websocket.Upgrader) *WsHub {
	return &WsHub{
		broadcast:  make(chan []byte),
		register:   make(chan *WsClient),
		unregister: make(chan *WsClient),
		clients:    make(map[*WsClient]bool),
		upgrader:   upgrader,
	}
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type WsHub struct {
	clients    map[*WsClient]bool // Registered clients.
	broadcast  chan []byte        // Inbound messages from the clients.
	register   chan *WsClient     // Register requests from the clients.
	unregister chan *WsClient     // Unregister requests from clients.
	upgrader   websocket.Upgrader
}

func (h *WsHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("websocket connection closed")
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
