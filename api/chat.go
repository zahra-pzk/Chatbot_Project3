package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zahra-pzk/Chatbot_Project3/api/ws"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
)

type createChatRequest struct {
	Content string `json:"content" binding:"required"`
	Name    string `json:"name"`
}

func (server *Server) createChat(ctx *gin.Context) {
	var req createChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateChatParams{
		UserExternalID: authPayload.UserExternalID,
		Status:         "pending",
	}

	chat, err := server.store.CreateChat(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	msgArg := db.CreateMessageParams{
		ChatExternalID:   chat.ChatExternalID,
		SenderExternalID: authPayload.UserExternalID,
		Content:          req.Content,
		IsSystemMessage:  false,
	}
	_, err = server.store.CreateMessage(ctx, msgArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	chatItem := struct {
		ChatExternalID string `json:"chat_external_id"`
		UserExternalID string `json:"user_external_id"`
		Status         string `json:"status"`
		UpdatedAt      string `json:"updated_at"`
	}{
		ChatExternalID: chat.ChatExternalID.String(),
		UserExternalID: chat.UserExternalID.String(),
		Status:         chat.Status,
		UpdatedAt:      chat.UpdatedAt.Time.Format(time.RFC3339),
	}
	jsonBytes, _ := json.Marshal([]interface{}{chatItem})
	server.hub.Broadcast <- ws.BroadcastMessage{
		ChatExternalID: adminChannelID,
		Data:           jsonBytes,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"chat_external_id": chat.ChatExternalID,
		"user_external_id": chat.UserExternalID,
		"status":           chat.Status,
		"created_at":       chat.CreatedAt,
	})
}

func (server *Server) getChat(ctx *gin.Context) {
	chatIdStr := ctx.Param("chatExternalID")
	chatUUID, err := uuid.Parse(chatIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	chat, err := server.store.GetChat(ctx, chatUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, chat)
}

func (server *Server) getChatsByUser(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.GetChatsByUserParams{
		UserExternalID: authPayload.UserExternalID,
		Limit:          10,
		Offset:         0,
	}
	chats, err := server.store.GetChatsByUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, chats)
}

func (server *Server) listChats(ctx *gin.Context) {
	arg := db.ListChatsParams{
		Limit:  10,
		Offset: 0,
	}
	chats, err := server.store.ListChats(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, chats)
}

func (server *Server) deleteChat(ctx *gin.Context) {
	chatIdStr := ctx.Param("chatExternalID")
	chatUUID, err := uuid.Parse(chatIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	err = server.store.DeleteChat(ctx, chatUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}

type updateChatStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open pending closed"`
}

func (server *Server) updateChatStatus(ctx *gin.Context) {
	chatIdStr := ctx.Param("chatExternalID")
	chatUUID, err := uuid.Parse(chatIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var req updateChatStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.UpdateChatStatusParams{
		ChatExternalID: chatUUID,
		Status:         req.Status,
	}
	chat, err := server.store.UpdateChatStatus(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, chat)
}

type updateChatRequest struct {
	Status string `json:"status"`
}

func (server *Server) updateChat(ctx *gin.Context) {
	chatIdStr := ctx.Param("chatExternalID")
	chatUUID, err := uuid.Parse(chatIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var req updateChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.UpdateChatParams{
		ChatExternalID: chatUUID,
		Status:         req.Status,
	}
	chat, err := server.store.UpdateChat(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, chat)
}
