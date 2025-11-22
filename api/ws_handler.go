package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSConnection struct {
	Conn           *websocket.Conn
	UserExternalID uuid.UUID
	ChatExternalID uuid.UUID
	Send           chan []byte
}

type WSHub struct {
	Clients    map[uuid.UUID]map[uuid.UUID]*WSConnection
	Register   chan *WSConnection
	Unregister chan *WSConnection
	Broadcast  chan WSMessage
}

type WSMessage struct {
	ChatExternalID uuid.UUID `json:"chat_external_id"`
	SenderUUID     uuid.UUID `json:"sender_uuid"`
	Content        string    `json:"content"`
}

var Hub = &WSHub{
	Clients:    make(map[uuid.UUID]map[uuid.UUID]*WSConnection),
	Register:   make(chan *WSConnection),
	Unregister: make(chan *WSConnection),
	Broadcast:  make(chan WSMessage),
}

func (server *Server) WSHandler(ctx *gin.Context) {
	chatIDStr := ctx.Query("chat_external_id")
	chatUUID, err := uuid.Parse(chatIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authPayload := ctx.MustGet("authorization_payload").(*token.Payload)
	userUUID := authPayload.UserExternalID

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	client := &WSConnection{
		Conn:           conn,
		UserExternalID: userUUID,
		ChatExternalID: chatUUID,
		Send:           make(chan []byte),
	}

	Hub.Register <- client

	go client.readPump(server)
	go client.writePump()
}

func (c *WSConnection) readPump(server *Server) {
	defer func() {
		Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var payload WSMessage
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}

		arg := db.CreateMessageParams{
			ChatExternalID:  c.ChatExternalID,
			SenderExternalID: c.UserExternalID,
			Content:         payload.Content,
			IsSystemMessage: false,
		}
		server.store.CreateMessage(context.Background(), arg)

		Hub.Broadcast <- payload
	}
}

func (c *WSConnection) writePump() {
	for msg := range c.Send {
		c.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func StartHub() {
	go func() {
		for {
			select {
			case client := <-Hub.Register:
				if Hub.Clients[client.ChatExternalID] == nil {
					Hub.Clients[client.ChatExternalID] = make(map[uuid.UUID]*WSConnection)
				}
				Hub.Clients[client.ChatExternalID][client.UserExternalID] = client
			case client := <-Hub.Unregister:
				if clients, ok := Hub.Clients[client.ChatExternalID]; ok {
					delete(clients, client.UserExternalID)
					close(client.Send)
				}
			case message := <-Hub.Broadcast:
				if clients, ok := Hub.Clients[message.ChatExternalID]; ok {
					for _, c := range clients {
						c.Send <- []byte(message.Content)
					}
				}
			}
		}
	}()
}
