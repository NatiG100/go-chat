package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"example.com/go-chat/internal/core"
	"example.com/go-chat/internal/core/usecases"
	"example.com/go-chat/internal/drivers"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// WSHandler is the entrypoint used in main: WSHandler(rds, jwt, repos)
func WSHandler(rds *drivers.RedisClient, jwt *drivers.JWTManager, repos core.Repositories) gin.HandlerFunc {
	hub := NewHub(rds)
	chatU := usecases.NewChatUsecase(repos, rds)

	go hub.Run(context.Background())

	return func(c *gin.Context) {
		// auth token via query or header
		tok := c.GetHeader("Authorization")
		if tok == "" {
			tok = c.Query("token")
		}
		if tok == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		if len(tok) > 7 && tok[:7] == "Bearer " {
			tok = tok[7:]
		}
		uid, err := jwt.Verify(tok)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("upgrade", err)
			return
		}

		client := &Client{hub: hub, conn: ws, send: make(chan []byte, 256), userID: uid}
		hub.Register(client)

		go client.writePump()
		go client.readPump(chatU)
	}
}

// Hub holds local clients and maps userID to clients. It also subscribes to redis channels for incoming messages.
type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	mu         sync.RWMutex
	rds        *drivers.RedisClient
	register   chan *Client
	unregister chan *Client
	ctx        context.Context
}

func NewHub(rds *drivers.RedisClient) *Hub {
	return &Hub{clients: make(map[uuid.UUID]map[*Client]bool), rds: rds, register: make(chan *Client), unregister: make(chan *Client)}
}

func (h *Hub) Run(ctx context.Context) {
	h.ctx = ctx
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[c.userID]; !ok {
				h.clients[c.userID] = make(map[*Client]bool)
			}
			h.clients[c.userID][c] = true
			h.mu.Unlock()
			go h.subscribePrivate(c.userID)
		case c := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[c.userID]; ok {
				delete(conns, c)
				if len(conns) == 0 {
					delete(h.clients, c.userID)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(c *Client)   { h.register <- c }
func (h *Hub) Unregister(c *Client) { h.unregister <- c }

func (h *Hub) subscribePrivate(userID uuid.UUID) {
	channel := "private:" + userID.String()
	ps := h.rds.Subscribe(h.ctx, channel)
	ch := ps.Channel()
	for msg := range ch {
		var m core.Message
		if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil {
			continue
		}
		h.mu.RLock()
		cons := h.clients[userID]
		for c := range cons {
			select {
			case c.send <- []byte(msg.Payload):
			default:
				// drop if blocked
			}
		}
		h.mu.RUnlock()
	}
}

// Client represents a ws connection
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID uuid.UUID
}

func (c *Client) readPump(chatU *usecases.ChatUsecase) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })
	for {
		var raw map[string]any
		if err := c.conn.ReadJSON(&raw); err != nil {
			break
		}
		typeStr, _ := raw["type"].(string)
		switch typeStr {
		case "private_message":
			toStr, _ := raw["to"].(string)
			content, _ := raw["content"].(string)
			if toStr == "" || content == "" {
				continue
			}
			toID, _ := uuid.Parse(toStr)
			m, err := chatU.SendPrivate(context.Background(), c.userID, toID, content)
			if err == nil {
				b, _ := json.Marshal(m)
				// send to sender locally (ack)
				select {
				case c.send <- b:
				default:
				}
			}
		case "group_message":
			gidStr, _ := raw["group_id"].(string)
			content, _ := raw["content"].(string)
			if gidStr == "" || content == "" {
				continue
			}
			gid, _ := uuid.Parse(gidStr)
			m, err := chatU.SendGroup(context.Background(), c.userID, gid, content)
			if err == nil {
				b, _ := json.Marshal(m)
				// local send (ack). Group distribution happens via redis subscription for members.
				select {
				case c.send <- b:
				default:
				}
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
