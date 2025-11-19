package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CreateChat(ctx context.Context, arg CreateChatParams) (Chat, error)
	CreateChatDefaults(ctx context.Context, userExternalID uuid.UUID) (Chat, error)
	DeleteChat(ctx context.Context, chatExternalID uuid.UUID) error
	GetChat(ctx context.Context, chatExternalID uuid.UUID) (Chat, error)
	GetChatsByUser(ctx context.Context, arg GetChatsByUserParams) ([]Chat, error)
	ListChats(ctx context.Context, arg ListChatsParams) ([]Chat, error)
	UpdateChat(ctx context.Context, arg UpdateChatParams) (Chat, error)
	UpdateChatStatus(ctx context.Context, arg UpdateChatStatusParams) (Chat, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteUser(ctx context.Context, userExternalID uuid.UUID) error
	GetUser(ctx context.Context, userExternalID uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email pgtype.Text) (User, error)
	GetUserByUsername(ctx context.Context, username pgtype.Text) (User, error)
	ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error
	CreateMessage(ctx context.Context, arg CreateMessageParams) (Message, error)
	DeleteMessage(ctx context.Context, messageExternalID uuid.UUID) error
	GetMessage(ctx context.Context, messageExternalID uuid.UUID) (Message, error)
	ListMessagesByChat(ctx context.Context, arg ListMessagesByChatParams) ([]Message, error)
	ListRecentMessagesByChat(ctx context.Context, arg ListRecentMessagesByChatParams) ([]Message, error)
	UpdateMessage(ctx context.Context, arg UpdateMessageParams) (Message, error)
}