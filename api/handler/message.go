package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zahra-pzk/Chatbot_Project3/api/dto"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

type MessageHandler struct {
	store      *db.SQLStore
	tokenMaker token.Maker
	config     util.Config
}

func NewMessageHandler(store *db.SQLStore, tokenMaker token.Maker, config util.Config) *MessageHandler {
	return &MessageHandler{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	payloadKey, exists := c.Get(authorizationPayloadKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, util.ErrorResponse(errors.New("authorization required")))
		return
	}
	payload := payloadKey.(*token.Payload)

	user, err := h.store.Querier.GetUserByExternalID(c, payload.UserExternalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	chatExternalID, err := uuid.Parse(req.ChatExternalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("invalid chat id")))
		return
	}

	isAdmin := user.Role == string(db.RoleTypeAdmin) || user.Role == string(db.RoleTypeSuperadmin)
	isSystem := user.Role == string(db.RoleTypeSystem)

	arg := db.CreateMessageParams{
		ChatExternalID:   chatExternalID,
		SenderExternalID: user.UserExternalID,
		Content:          req.Content,
		IsSystemMessage:  isSystem,
		IsAdminMessage:   isAdmin,
	}

	msg, err := h.store.Querier.CreateMessage(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	var attachments []dto.Attachment
	for _, url := range req.AttachmentURLs {
		attachArg := db.CreateAttachmentParams{
			MessageExternalID: msg.MessageExternalID,
			UserExternalID:    user.UserExternalID,
			ChatExternalID:    chatExternalID,
			Url:               url,
			Filename:          pgtype.Text{String: "unknown", Valid: true},
			MimeType:          pgtype.Text{String: "application/octet-stream", Valid: true},
			SizeBytes:         pgtype.Int8{Int64: 0, Valid: true},
			Metadata:          []byte("{}"),
		}
		newAttach, err := h.store.Querier.CreateAttachment(c, attachArg)
		if err == nil {
			attachments = append(attachments, dto.Attachment{
				AttachmentExternalID: newAttach.AttachmentExternalID.String(),
				URL:                  newAttach.Url,
				Filename:             newAttach.Filename.String,
				MimeType:             newAttach.MimeType.String,
				SizeBytes:            newAttach.SizeBytes.Int64,
				CreatedAt:            newAttach.CreatedAt.Time,
			})
		}
	}

	rsp := dto.MessageResponse{
		MessageExternalID: msg.MessageExternalID.String(),
		ChatExternalID:    msg.ChatExternalID.String(),
		SenderExternalID:  msg.SenderExternalID.String(),
		Content:           msg.Content,
		IsSystemMessage:   msg.IsSystemMessage,
		IsAdminMessage:    msg.IsAdminMessage,
		CreatedAt:         msg.CreatedAt.Time,
		UpdatedAt:         msg.UpdatedAt.Time,
		Attachments:       attachments,
		Reactions:         []dto.Reaction{},
	}

	c.JSON(http.StatusCreated, rsp)
}

func (h *MessageHandler) ListMessages(c *gin.Context) {
	var req dto.MessageHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	chatExternalIDStr := c.Param("chat_id")
	chatExternalID, err := uuid.Parse(chatExternalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("invalid chat id")))
		return
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	arg := db.ListMessagesByChatParams{
		ChatExternalID: chatExternalID,
		Limit:          req.Limit,
		Offset:         req.Offset,
	}

	messages, err := h.store.Querier.ListMessagesByChat(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	var rsp []dto.MessageResponse
	for _, m := range messages {
		attachments, _ := h.store.Querier.ListAllAttachmentsByMessage(c, m.MessageExternalID)
		reactions, _ := h.store.Querier.ListAllReactionsByMessage(c, m.MessageExternalID)

		var attachDTOs []dto.Attachment
		for _, a := range attachments {
			attachDTOs = append(attachDTOs, dto.Attachment{
				AttachmentExternalID: a.AttachmentExternalID.String(),
				URL:                  a.Url,
				Filename:             a.Filename.String,
				MimeType:             a.MimeType.String,
				SizeBytes:            a.SizeBytes.Int64,
				CreatedAt:            a.CreatedAt.Time,
			})
		}
		if attachDTOs == nil {
			attachDTOs = []dto.Attachment{}
		}

		var reactDTOs []dto.Reaction
		for _, r := range reactions {
			reactDTOs = append(reactDTOs, dto.Reaction{
				ReactionExternalID: r.ReactionExternalID.String(),
				UserExternalID:     r.UserExternalID.String(),
				Reaction:           r.Reaction,
				Score:              r.Score,
				CreatedAt:          r.CreatedAt.Time,
			})
		}
		if reactDTOs == nil {
			reactDTOs = []dto.Reaction{}
		}

		rsp = append(rsp, dto.MessageResponse{
			MessageExternalID: m.MessageExternalID.String(),
			ChatExternalID:    m.ChatExternalID.String(),
			SenderExternalID:  m.SenderExternalID.String(),
			Content:           m.Content,
			IsSystemMessage:   m.IsSystemMessage,
			IsAdminMessage:    m.IsAdminMessage,
			CreatedAt:         m.CreatedAt.Time,
			UpdatedAt:         m.UpdatedAt.Time,
			Attachments:       attachDTOs,
			Reactions:         reactDTOs,
		})
	}

	if rsp == nil {
		rsp = []dto.MessageResponse{}
	}

	c.JSON(http.StatusOK, rsp)
}

func (h *MessageHandler) EditMessage(c *gin.Context) {
	var req dto.EditMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	payloadKey, _ := c.Get(authorizationPayloadKey)
	payload := payloadKey.(*token.Payload)

	msgIDStr := c.Param("id")
	msgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("invalid message id")))
		return
	}

	msg, err := h.store.Querier.GetMessage(c, msgID)
	if err != nil {
		c.JSON(http.StatusNotFound, util.ErrorResponse(err))
		return
	}

	if payload.Role != string(db.RoleTypeSuperadmin) {
		if payload.Role == string(db.RoleTypeAdmin) {
			if msg.SenderExternalID != payload.UserExternalID {
				c.JSON(http.StatusForbidden, util.ErrorResponse(errors.New("admin can only edit their own messages")))
				return
			}
		} else {
			c.JSON(http.StatusForbidden, util.ErrorResponse(errors.New("permission denied")))
			return
		}
	}

	arg := db.EditMessageParams{
		MessageExternalID: msgID,
		Content:           req.NewContent,
	}

	updatedMsg, err := h.store.Querier.EditMessage(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, mapEditMessageToDTO(updatedMsg))
}

func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	msgIDStr := c.Param("id")
	msgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("invalid message id")))
		return
	}

	payloadKey, _ := c.Get(authorizationPayloadKey)
	payload := payloadKey.(*token.Payload)

	msg, err := h.store.Querier.GetMessage(c, msgID)
	if err != nil {
		c.JSON(http.StatusNotFound, util.ErrorResponse(err))
		return
	}

	if payload.Role != string(db.RoleTypeSuperadmin) {
		if payload.Role == string(db.RoleTypeAdmin) {
			if msg.SenderExternalID != payload.UserExternalID {
				c.JSON(http.StatusForbidden, util.ErrorResponse(errors.New("admin can only delete their own messages")))
				return
			}
		} else {
			c.JSON(http.StatusForbidden, util.ErrorResponse(errors.New("permission denied")))
			return
		}
	}

	err = h.store.Querier.DeleteMessage(c, msgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "message deleted"})
}

func (h *MessageHandler) DeleteMessagesByChat(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("invalid chat id")))
		return
	}

	if err := h.store.Querier.DeleteMessagesByChat(c, chatID); err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "all messages in chat deleted"})
}

func mapEditMessageToDTO(m db.EditMessageRow) dto.MessageResponse {
	return dto.MessageResponse{
		MessageExternalID: m.MessageExternalID.String(),
		ChatExternalID:    m.ChatExternalID.String(),
		SenderExternalID:  m.SenderExternalID.String(),
		Content:           m.Content,
		IsSystemMessage:   m.IsSystemMessage,
		IsAdminMessage:    m.IsAdminMessage,
		CreatedAt:         m.CreatedAt.Time,
		UpdatedAt:         m.UpdatedAt.Time,
		Attachments:       []dto.Attachment{},
		Reactions:         []dto.Reaction{},
	}
}