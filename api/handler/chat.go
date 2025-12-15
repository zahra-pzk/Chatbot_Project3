package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zahra-pzk/Chatbot_Project3/api/dto"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

type ChatHandler struct {
	store      *db.SQLStore
	tokenMaker token.Maker
	config     util.Config
}

func NewChatHandler(store *db.SQLStore, tokenMaker token.Maker, config util.Config) *ChatHandler {
	return &ChatHandler{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}
}

func (h *ChatHandler) StartChat(c *gin.Context) {
	var req dto.CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	var userExternalID uuid.UUID
	var accessToken string
	var userEmail string
	var userName string

	payloadKey, exists := c.Get(authorizationPayloadKey)

	if exists {
		payload := payloadKey.(*token.Payload)
		userExternalID = payload.UserExternalID
		user, err := h.store.Querier.GetUserByExternalID(c, userExternalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
			return
		}
		userEmail = user.Email
		userName = user.FirstName + " " + user.LastName
	} else {
		if req.Email == "" || req.FirstName == "" || req.LastName == "" {
			c.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("guest details required")))
			return
		}

		arg := db.CreateGuestUserParams{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Role:      string(db.RoleTypeGuest),
		}

		user, err := h.store.Querier.CreateGuestUser(c, arg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
			return
		}

		userExternalID = user.UserExternalID
		userEmail = user.Email
		userName = user.FirstName + " " + user.LastName

		token, _, err := h.tokenMaker.CreateToken(
			user.UserExternalID,
			user.Username.String,
			user.Role,
			h.config.AccessTokenDuration,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
			return
		}
		accessToken = token
	}

	chatLabel := req.Label
	if chatLabel == "" {
		chatLabel = "Empty Label"
	}

	chatArg := db.CreateChatParams{
		UserExternalID:  userExternalID,
		Column2:         string(db.ChatStatusTypeOpen),
		Label:           chatLabel,
		AdminExternalID: pgtype.UUID{Valid: false},
		Score:           pgtype.Int8{Int64: 0, Valid: true},
	}

	chat, err := h.store.Querier.CreateChat(c, chatArg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	fmt.Printf("NOTIFICATION: New chat started by %s (%s) - ChatID: %s\n", userName, userEmail, chat.ChatExternalID)

	rsp := dto.CreateChatResponse{
		ChatID:         chat.ChatExternalID.String(),
		UserExternalID: chat.UserExternalID.String(),
		Label:          chat.Label,
		Status:         string(chat.Status),
		AccessToken:    accessToken,
		CreatedAt:      chat.CreatedAt.Time,
	}

	c.JSON(http.StatusOK, rsp)
}

func (h *ChatHandler) CloseChat(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	payload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	chat, err := h.store.Querier.GetChat(c, chatID)
	if err != nil {
		c.JSON(http.StatusNotFound, util.ErrorResponse(err))
		return
	}

	if payload.Role != string(db.RoleTypeSuperadmin) && payload.Role != string(db.RoleTypeAdmin) {
		if chat.UserExternalID != payload.UserExternalID {
			c.JSON(http.StatusForbidden, util.ErrorResponse(errors.New("permission denied")))
			return
		}
	}

	arg := db.UpdateChatStatusParams{
		ChatExternalID: chatID,
		Column2:        string(db.ChatStatusTypeClosed),
	}

	_, err = h.store.Querier.UpdateChatStatus(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "chat closed"})
}

func (h *ChatHandler) DeleteChat(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if err := h.store.Querier.DeleteChat(c, chatID); err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "chat deleted"})
}
