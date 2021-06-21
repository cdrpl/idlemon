package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WsClient struct {
	conn   *websocket.Conn
	send   chan WebSocketMessage // Buffered channel of outbound messages.
	userId uuid.UUID
	wsHub  *WsHub
}

func (c *WsClient) readPump() {
	defer func() {
		c.wsHub.unregisterClient <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(WS_MAX_MESSAGE_SIZE)
	c.conn.SetReadDeadline(time.Now().Add(WS_PONG_TIMEOUT))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(WS_PONG_TIMEOUT)); return nil })

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *WsClient) writePump() {
	ticker := time.NewTicker(WS_PING_INTERVAL)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(WS_WRITE_TIMOUT))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
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

func WsConnectHandler(ctx context.Context, rdb *redis.Client, wsHub *WsHub, w http.ResponseWriter, r *http.Request) {
	id, token := ParseAuthHeader(r.Header.Get("Authorization"))

	authenticated, err := ValidateApiToken(r.Context(), id, token, rdb)
	if err != nil {
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	if authenticated {
		userId, err := uuid.Parse(id)
		if err != nil {
			log.Printf("fail to parse user ID: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}

		conn, err := wsHub.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("fail to upgrade WebSocket connection: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}

		client := &WsClient{wsHub: wsHub, userId: userId, conn: conn, send: make(chan WebSocketMessage, 256)}
		wsHub.registerClient <- client

		go client.writePump()
		go client.readPump()
	}
}

// WsHub maintains the set of active clients and broadcasts messages to the clients.
type WsHub struct {
	clients          map[uuid.UUID]*WsClient // Key is the user's ID.
	registerClient   chan *WsClient
	unregisterClient chan *WsClient
	broadcast        chan WebSocketMessage
	shutdown         chan bool
	upgrader         websocket.Upgrader
}

func CreateWsHub(upgrader websocket.Upgrader) *WsHub {
	return &WsHub{
		clients:          make(map[uuid.UUID]*WsClient),
		registerClient:   make(chan *WsClient),
		unregisterClient: make(chan *WsClient),
		broadcast:        make(chan WebSocketMessage),
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

			h.clients[client.userId] = client
			log.Printf("user %v WebSocket connected\n", client.userId)

		case client := <-h.unregisterClient:
			if client, ok := h.clients[client.userId]; ok {
				delete(h.clients, client.userId)
				close(client.send)
				log.Printf("user %v WebSocket closed\n", client.userId)
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

			h.shutdown <- true // signifies shutdown complete
		}
	}
}

func (h *WsHub) AuthenticateClient(client *WsClient) bool {
	return true
}

type WebSocketMessage struct {
	Type int    `json:"type"`
	Data []byte `json:"data"`
}

type WebSocketChatMessage struct {
	SenderName string `json:"senderName"`
	Message    string `json:"message"`
}

func CreateWebSocketChatMessage(senderName string, message string) WebSocketMessage {
	chatMsg := WebSocketChatMessage{
		SenderName: senderName,
		Message:    message,
	}

	bytes, err := json.Marshal(chatMsg)
	if err != nil {
		log.Printf("fail to marshall chat message: %v\n", err)
	}

	wsMsg := WebSocketMessage{
		Type: WS_CHAT_MESSAGE,
		Data: bytes,
	}

	return wsMsg
}
