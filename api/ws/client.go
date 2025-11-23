package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	Hub            *Hub
	Conn           *websocket.Conn
	Send           chan []byte
	Store          *db.SQLStore
	ChatExternalID uuid.UUID
	UserExternalID uuid.UUID
	Role           db.RoleType
}

func (c *Client) ReadPump() {
	defer func() { c.Hub.Unregister <- c; c.Conn.Close() }()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			} else {
				log.Printf("error in reading msg: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(message)

		arg := db.CreateMessageParams{
			ChatExternalID:   c.ChatExternalID,
			SenderExternalID: c.UserExternalID,
			Content:          string(message),
		}

		msg, err := c.Store.CreateMessage(context.Background(), arg)
		if err != nil {
			log.Printf("error creating message: %v", err)
			continue
		}
		jsonBytes, _ := json.Marshal(msg)
		c.Hub.Broadcast <- BroadcastMessage{ChatExternalID: c.ChatExternalID, Data: jsonBytes}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
