package dto

import "time"

type SendMessageRequest struct {
	ChatExternalID string   `json:"chat_external_id"`
	Content        string   `json:"content"`
	AttachmentURLs []string `json:"attachment_urls,omitempty"`
}

type MessageHistoryRequest struct {
	Limit  int32 `form:"limit"`
	Offset int32 `form:"offset"`
}

type MessageResponse struct {
	MessageExternalID string       `json:"message_external_id"`
	ChatExternalID    string       `json:"chat_external_id"`
	SenderExternalID  string       `json:"sender_external_id"`
	Content           string       `json:"content"`
	IsSystemMessage   bool         `json:"is_system_message"`
	IsAdminMessage    bool         `json:"is_admin_message"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	Attachments       []Attachment `json:"attachments"`
	Reactions         []Reaction   `json:"reactions"`
}


type EditMessageRequest struct {
	MessageExternalID string `json:"message_external_id"`
	NewContent        string `json:"new_content"`
}

type DeleteMessageRequest struct {
	MessageExternalID string `json:"message_external_id"`
}

type TypingEvent struct {
	ChatExternalID string `json:"chat_external_id"`
	IsTyping       bool   `json:"is_typing"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}