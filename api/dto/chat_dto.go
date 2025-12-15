package dto

import (
	"time"
)

type CreateChatRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Label     string `json:"label"`
}

type UpdateChatRequest struct {
	Label  string `json:"label"`
	Status string `json:"status"`
}

type CreateChatResponse struct {
	ChatID         string    `json:"chat_id"`
	UserExternalID string    `json:"user_external_id"`
	Label          string    `json:"label"`
	Status         string    `json:"status"`
	AccessToken    string    `json:"access_token,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type CloseChatRequest struct {
	ChatID string `json:"chat_id" binding:"required"`
}

type ChatResponse struct {
	ChatExternalID  string    `json:"chat_external_id"`
	UserExternalID  string    `json:"user_external_id"`
	AdminExternalID *string   `json:"admin_external_id,omitempty"`
	Label           string    `json:"label"`
	Status          string    `json:"status"`
	Score           int64     `json:"score"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	UnreadCount     int       `json:"unread_count,omitempty"`
	LastMessage     string    `json:"last_message,omitempty"`
}

type GetChatsRequest struct {
	Limit  int32 `form:"limit"`
	Offset int32 `form:"offset"`
}