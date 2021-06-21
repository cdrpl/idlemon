package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	conn          *websocket.Conn
	send          chan []byte // Buffered channel of outbound messages.
	userId        int
	authenticated bool
	wsHub         *WsHub
}

func (c *WsClient) readPump() {
	defer func() {
		if c.authenticated {
			c.wsHub.unregisterClient <- c
		} else {
			c.wsHub.unregisterAnon <- c
		}
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
	wsHub.registerAnon <- client

	go client.writePump()
	go client.readPump()

	// close anon connection if it doesn't authenticate in time
	go func() {
		time.Sleep(WS_AUTH_TIMEOUT)
		wsHub.unregisterAnon <- client
	}()
}

// WsHub maintains the set of active clients and broadcasts messages to the clients.
type WsHub struct {
	clients          map[int]*WsClient
	anonClients      map[*WsClient]bool
	registerClient   chan *WsClient
	registerAnon     chan *WsClient
	unregisterClient chan *WsClient
	unregisterAnon   chan *WsClient
	broadcast        chan []byte
	shutdown         chan bool
	upgrader         websocket.Upgrader
}

func CreateWsHub(upgrader websocket.Upgrader) *WsHub {
	return &WsHub{
		clients:          make(map[int]*WsClient),
		anonClients:      make(map[*WsClient]bool),
		registerClient:   make(chan *WsClient),
		registerAnon:     make(chan *WsClient),
		unregisterClient: make(chan *WsClient),
		unregisterAnon:   make(chan *WsClient),
		broadcast:        make(chan []byte),
		shutdown:         make(chan bool),
		upgrader:         upgrader,
	}
}

func (h *WsHub) Run() {
	for {
		select {
		case client := <-h.registerClient:
			// close current authenticated client if exists
			if client, ok := h.clients[client.userId]; ok {
				close(client.send)
			}

			// add client to authenticated clients and remove from anon clients
			h.clients[client.userId] = client
			delete(h.anonClients, client)

			log.Printf("user %v authenticated WebSocket connection\n", client.userId)

		case client := <-h.registerAnon:
			h.anonClients[client] = true
			log.Println("anonymous websocket client connected")

		case client := <-h.unregisterClient:
			if client, ok := h.clients[client.userId]; ok {
				delete(h.clients, client.userId)
				close(client.send)
				log.Printf("user %v websocket connection closed\n", client.userId)
			}

		case client := <-h.unregisterAnon:
			if _, ok := h.anonClients[client]; ok {
				delete(h.anonClients, client)
				close(client.send)
				log.Println("anonymous websocket connection closed")
			}

		case msg := <-h.broadcast:
			for userId, client := range h.clients {
				select {
				case client.send <- msg:

				default:
					close(client.send)
					delete(h.clients, userId)
				}
			}

		case <-h.shutdown:
			for userId, client := range h.clients {
				delete(h.clients, userId)
				close(client.send)
			}

			for client := range h.anonClients {
				close(client.send)
				delete(h.anonClients, client)
			}

			h.shutdown <- true // signifies shutdown complete
		}
	}
}

func (h *WsHub) AuthenticateClient(client *WsClient) bool {
	return true
}
