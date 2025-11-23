package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zahra-pzk/Chatbot_Project3/api/ws"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (server *Server) ServeWs(ctx *gin.Context) {
	chatIdStr := ctx.Param("chatExternalID")
	chatUUID, err := uuid.Parse(chatIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("token is required")))
		return
	}

	payload, err := server.tokenMaker.VerifyToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByExternalID(ctx, payload.UserExternalID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	chat, err := server.store.GetChat(ctx, chatUUID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	userRole := db.RoleType(user.Role)

	if userRole == db.RoleTypeUser {
		if chat.UserExternalID != user.UserExternalID {
			ctx.JSON(http.StatusForbidden, errorResponse(fmt.Errorf("access denied")))
			return
		}
	} else if userRole == db.RoleTypeAdmin || userRole == db.RoleTypeSuperadmin {
		arg := db.UpdateChatStatusParams{
			ChatExternalID: chatUUID,
			Status:     "open",
		}
		server.store.UpdateChatStatus(ctx, arg)
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		Hub:            server.hub,
		Conn:           conn,
		Send:           make(chan []byte, 256),
		Store:          server.store,
		ChatExternalID: chatUUID,
		UserExternalID: payload.UserExternalID,
		Role:           userRole,
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}