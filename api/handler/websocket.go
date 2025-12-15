package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zahra-pzk/Chatbot_Project3/api/ws"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	store      *db.SQLStore
	tokenMaker token.Maker
	config     util.Config
	hub        *ws.Hub
}

func NewWebSocketHandler(store *db.SQLStore, tokenMaker token.Maker, config util.Config, hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
		hub:        hub,
	}
}

func (h *WebSocketHandler) ServeWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	chatIDStr := c.Param("id")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		log.Println("invalid chat id:", err)
		conn.Close()
		return
	}

	tokenString := c.Query("token")
	if tokenString == "" {
		log.Println("token not provided")
		conn.Close()
		return
	}

	payload, err := h.tokenMaker.VerifyToken(tokenString)
	if err != nil {
		log.Println("invalid token:", err)
		conn.Close()
		return
	}

	client := &ws.Client{
		Hub:            h.hub,
		Conn:           conn,
		Send:           make(chan []byte, 256),
		Store:          h.store,
		ChatExternalID: chatID,
		UserExternalID: payload.UserExternalID,
		Role:           payload.Role,
	}

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}