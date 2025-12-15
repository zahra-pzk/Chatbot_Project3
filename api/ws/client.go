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
	maxMessageSize = 8192
)

type Client struct {
	Hub            *Hub
	Conn           *websocket.Conn
	Send           chan []byte
	Store          *db.SQLStore
	ChatExternalID uuid.UUID
	UserExternalID uuid.UUID
	Role           string
}

type IncomingMessage struct {
	Content string `json:"content"`
}

type OutgoingMessage struct {
	Content          string    `json:"content"`
	SenderExternalID uuid.UUID `json:"sender_external_id"`
	CreatedAt        string    `json:"created_at"`
	IsSystem         bool      `json:"is_system"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
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
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(message)

		var incomingMsg IncomingMessage
		if err := json.Unmarshal(message, &incomingMsg); err != nil {
			continue
		}

		if c.Role == "admin" || c.Role == "superadmin" || c.Role == "system" {
			chat, _ := c.Store.Querier.GetChat(context.Background(), c.ChatExternalID)
			if chat.Status == "pending" {
				c.Store.Querier.UpdateChatStatus(context.Background(), db.UpdateChatStatusParams{
					ChatExternalID: c.ChatExternalID,
					Column2:        string(db.ChatStatusTypeOpen),
				})
			}
		}

		arg := db.CreateMessageParams{
			ChatExternalID:   c.ChatExternalID,
			SenderExternalID: c.UserExternalID,
			Content:          incomingMsg.Content,
			IsSystemMessage:  c.Role == "system",
			IsAdminMessage:   c.Role == "admin" || c.Role == "superadmin",
		}

		msg, err := c.Store.Querier.CreateMessage(context.Background(), arg)
		if err != nil {
			log.Printf("DB error: %v", err)
			continue
		}

		outMsg := OutgoingMessage{
			Content:          msg.Content,
			SenderExternalID: msg.SenderExternalID,
			CreatedAt:        msg.CreatedAt.Time.Format(time.RFC3339),
			IsSystem:         msg.IsSystemMessage,
		}

		jsonBytes, _ := json.Marshal(outMsg)

		c.Hub.Broadcast <- BroadcastMessage{
			ChatExternalID: c.ChatExternalID,
			Data:           jsonBytes,
		}
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
