package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
)

type createMessageRequest struct {
	ChatExternalID string `json:"chat_external_id" binding:"required"`
	Content        string `json:"content" binding:"required"`
}

func (server *Server) createMessage(ctx *gin.Context) {
	var req createMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	chatUUID, err := uuid.Parse(req.ChatExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.CreateMessageParams{
		ChatExternalID:   chatUUID,
		SenderExternalID: authPayload.UserExternalID,
		Content:          req.Content,
		IsSystemMessage:  false,
	}
	msg, err := server.store.CreateMessage(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, msg)
}

func (server *Server) listMessagesByChat(ctx *gin.Context) {
	chatIdStr := ctx.Param("chatExternalID")
	chatUUID, err := uuid.Parse(chatIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListMessagesByChatParams{
		ChatExternalID: chatUUID,
		Limit:          100,
		Offset:         0,
	}
	msgs, err := server.store.ListMessagesByChat(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, msgs)
}

func (server *Server) listRecentMessagesByChat(ctx *gin.Context) {
	chatIdStr := ctx.Param("chatExternalID")
	chatUUID, err := uuid.Parse(chatIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListRecentMessagesByChatParams{
		ChatExternalID: chatUUID,
		Limit:          20,
	}
	msgs, err := server.store.ListRecentMessagesByChat(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, msgs)
}

func (server *Server) getMessage(ctx *gin.Context) {
	msgIdStr := ctx.Param("messageExternalID")
	msgUUID, err := uuid.Parse(msgIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	msg, err := server.store.GetMessage(ctx, msgUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, msg)
}

type updateMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

func (server *Server) updateMessage(ctx *gin.Context) {
	msgIdStr := ctx.Param("messageExternalID")
	msgUUID, err := uuid.Parse(msgIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var req updateMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.UpdateMessageParams{
		MessageExternalID: msgUUID,
		Content:           req.Content,
	}
	msg, err := server.store.UpdateMessage(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, msg)
}

func (server *Server) deleteMessage(ctx *gin.Context) {
	msgIdStr := ctx.Param("messageExternalID")
	msgUUID, err := uuid.Parse(msgIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	err = server.store.DeleteMessage(ctx, msgUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}
