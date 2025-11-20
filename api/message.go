package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
)


type createMessageRequest struct {
	ChatExternalID  string  `json:"chat_external_id" binding:"required"`
	SenderExternalID string  `json:"sender_external_id" binding:"required"`
	Content     string  `json:"content" binding:"required"`
	IsSystemMessage bool   `json:"is_system_message"`
}

type messageIDRequest struct {
	MessageExternalID string `uri:"messageExternalID" binding:"required"`
}

type listMessagesByChatRequest struct {
	ChatExternalID string `uri:"chatExternalID" binding:"required"`
	PageID    int32  `form:"page_id" binding:"required,min=1"`
	PageSize   int32  `form:"page_size" binding:"required,min=5,max=50"`
}

type listRecentMessagesRequest struct {
	ChatExternalID string `uri:"chatExternalID" binding:"required"`
	Limit     int32  `form:"limit" binding:"required,min=1,max=50"`
}


type updateMessageBodyRequest struct {
	Content string `json:"content" binding:"required"`
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
    senderUUID, err := uuid.Parse(req.SenderExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateMessageParams{
		ChatExternalID:  chatUUID,
		SenderExternalID: senderUUID,
		Content:     req.Content,
		IsSystemMessage: req.IsSystemMessage,
	}

	message, err := server.store.CreateMessage(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, message)
}

func (server *Server) getMessage(ctx *gin.Context) {
	var req messageIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	messageUUID, err := uuid.Parse(req.MessageExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	message, err := server.store.GetMessage(ctx, messageUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, message)
}

func (server *Server) listMessagesByChat(ctx *gin.Context) {
	var uriReq struct { ChatExternalID string `uri:"chatExternalID" binding:"required"` }
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var queryReq listMessagesByChatRequest
	if err := ctx.ShouldBindQuery(&queryReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
    
    chatUUID, err := uuid.Parse(uriReq.ChatExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}


	arg := db.ListMessagesByChatParams{
		ChatExternalID: chatUUID,
		Limit:     queryReq.PageSize,
		Offset:     (queryReq.PageID - 1) * queryReq.PageSize,
	}

	messages, err := server.store.ListMessagesByChat(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, messages)
}

func (server *Server) listRecentMessagesByChat(ctx *gin.Context) {
	var uriReq struct { ChatExternalID string `uri:"chatExternalID" binding:"required"` }
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var queryReq listRecentMessagesRequest
	if err := ctx.ShouldBindQuery(&queryReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
    
    chatUUID, err := uuid.Parse(uriReq.ChatExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListRecentMessagesByChatParams{
		ChatExternalID: chatUUID,
		Limit:     queryReq.Limit,
	}

	messages, err := server.store.ListRecentMessagesByChat(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, messages)
}

func (server *Server) updateMessage(ctx *gin.Context) {
	var uriReq messageIDRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	messageUUID, err := uuid.Parse(uriReq.MessageExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var bodyReq updateMessageBodyRequest
	if err := ctx.ShouldBindJSON(&bodyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateMessageParams{
		MessageExternalID: messageUUID,
		Content:      bodyReq.Content,
	}

	message, err := server.store.UpdateMessage(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, message)
}

func (server *Server) deleteMessage(ctx *gin.Context) {
	var req messageIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	messageUUID, err := uuid.Parse(req.MessageExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.store.DeleteMessage(ctx, messageUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}