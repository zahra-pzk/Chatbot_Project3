package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

	authHeader := ctx.GetHeader("Authorization")
	var tokenStr string
	if len(authHeader) > 0 {
		fields := strings.Fields(authHeader)
		if len(fields) == 2 && strings.ToLower(fields[0]) == "bearer" {
			tokenStr = fields[1]
		}
	}
	if tokenStr == "" {
		tokenStr = ctx.Query("token")
	}

	if tokenStr == "" {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("token is required")))
		return
	}

	payload, err := server.tokenMaker.VerifyToken(tokenStr)
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

	if userRole == db.RoleTypeUser || userRole == db.RoleTypeGuest {
		if chat.UserExternalID != user.UserExternalID {
			ctx.JSON(http.StatusForbidden, errorResponse(fmt.Errorf("access denied")))
			return
		}
	} else if userRole == db.RoleTypeAdmin || userRole == db.RoleTypeSuperadmin {
		if chat.Status == "pending" {
			arg := db.UpdateChatStatusParams{
				ChatExternalID: chatUUID,
				Status:         "open",
			}
			updatedChat, err := server.store.UpdateChatStatus(ctx, arg)
			if err == nil {
				chatItem := struct {
					ChatExternalID string `json:"chat_external_id"`
					UserExternalID string `json:"user_external_id"`
					Status         string `json:"status"`
					UpdatedAt      string `json:"updated_at"`
				}{
					ChatExternalID: updatedChat.ChatExternalID.String(),
					UserExternalID: updatedChat.UserExternalID.String(),
					Status:         updatedChat.Status,
					UpdatedAt:      updatedChat.UpdatedAt.Time.Format(time.RFC3339),
				}
				jsonBytes, _ := json.Marshal([]interface{}{chatItem})
				server.hub.Broadcast <- ws.BroadcastMessage{
					ChatExternalID: adminChannelID,
					Data:           jsonBytes,
				}
			}
		}
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

func (server *Server) ServeAdminChatsWs(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	var tokenStr string
	if len(authHeader) > 0 {
		fields := strings.Fields(authHeader)
		if len(fields) == 2 && strings.ToLower(fields[0]) == "bearer" {
			tokenStr = fields[1]
		}
	}
	if tokenStr == "" {
		tokenStr = ctx.Query("token")
	}

	if tokenStr == "" {
		return
	}

	payload, err := server.tokenMaker.VerifyToken(tokenStr)
	if err != nil {
		return
	}

	arg := db.ListChatsParams{
		Limit:  100,
		Offset: 0,
	}
	chats, err := server.store.ListChats(ctx, arg)
	if err == nil {
		var chatList []interface{}
		for _, c := range chats {
			chatList = append(chatList, struct {
				ChatExternalID string `json:"chat_external_id"`
				UserExternalID string `json:"user_external_id"`
				Status         string `json:"status"`
				UpdatedAt      string `json:"updated_at"`
			}{
				ChatExternalID: c.ChatExternalID.String(),
				UserExternalID: c.UserExternalID.String(),
				Status:         c.Status,
				UpdatedAt:      c.UpdatedAt.Time.Format(time.RFC3339),
			})
		}
		jsonBytes, _ := json.Marshal(chatList)
		conn.WriteMessage(websocket.TextMessage, jsonBytes)
	}

	client := &ws.Client{
		Hub:            server.hub,
		Conn:           conn,
		Send:           make(chan []byte, 256),
		Store:          server.store,
		ChatExternalID: adminChannelID,
		UserExternalID: payload.UserExternalID,
		Role:           db.RoleTypeAdmin,
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}