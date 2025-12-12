package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	// User
	AddPhotoToUserProfile(ctx context.Context, arg AddPhotoToUserProfileParams) (User, error)
	AddPhotosToUserProfile(ctx context.Context, arg AddPhotosToUserProfileParams) (User, error)
	CreateGuestUser(ctx context.Context, arg CreateGuestUserParams) (CreateGuestUserRow, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error)
	DeleteUser(ctx context.Context, userExternalID uuid.UUID) error
	GetUser(ctx context.Context, userExternalID uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByExternalID(ctx context.Context, userExternalID uuid.UUID) (User, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber pgtype.Text) (User, error)
	GetUserByUsername(ctx context.Context, username pgtype.Text) (User, error)
	ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error
	UpdateUserRole(ctx context.Context, arg UpdateUserRoleParams) (User, error)
	UpdateUserStatus(ctx context.Context, arg UpdateUserStatusParams) (User, error)

	// Chat
	AssignedAdminToChat(ctx context.Context, arg AssignedAdminToChatParams) (Chat, error)
	CountUserChats(ctx context.Context, userExternalID uuid.UUID) (int64, error)
	CreateChat(ctx context.Context, arg CreateChatParams) (Chat, error)
	CreateChatDefaults(ctx context.Context, arg CreateChatDefaultsParams) (Chat, error)
	DeleteChat(ctx context.Context, chatExternalID uuid.UUID) error
	GetChat(ctx context.Context, chatExternalID uuid.UUID) (Chat, error)
	GetChatsByAdmin(ctx context.Context, arg GetChatsByAdminParams) ([]Chat, error)
	GetChatsByStatusAndScoreRange(ctx context.Context, arg GetChatsByStatusAndScoreRangeParams) ([]Chat, error)
	GetChatsByUser(ctx context.Context, arg GetChatsByUserParams) ([]Chat, error)
	GetOpenChatByUser(ctx context.Context, userExternalID uuid.UUID) (Chat, error)
	GetPendingChatByUser(ctx context.Context, userExternalID uuid.UUID) (Chat, error)
	GetClosedChatByUser(ctx context.Context, userExternalID uuid.UUID) (Chat, error)
	GetTopChatsByScore(ctx context.Context, arg GetTopChatsByScoreParams) ([]Chat, error)
	ListChats(ctx context.Context, arg ListChatsParams) ([]Chat, error)
	ListClosedChats(ctx context.Context, arg ListClosedChatsParams) ([]Chat, error)
	ListOpenChats(ctx context.Context, arg ListOpenChatsParams) ([]Chat, error)
	ListPendingChats(ctx context.Context, arg ListPendingChatsParams) ([]Chat, error)
	UpdateChat(ctx context.Context, arg UpdateChatParams) (Chat, error)
	UpdateChatScore(ctx context.Context, arg UpdateChatScoreParams) (Chat, error)
	UpdateChatStatus(ctx context.Context, arg UpdateChatStatusParams) (Chat, error)

	// Message
	CreateMessage(ctx context.Context, arg CreateMessageParams) (CreateMessageRow, error)
	EditMessage(ctx context.Context, arg EditMessageParams) (EditMessageRow, error)
	DeleteMessage(ctx context.Context, messageExternalID uuid.UUID) error
	DeleteMessagesByChat(ctx context.Context, chatExternalID uuid.UUID) error
	GetMessage(ctx context.Context, messageExternalID uuid.UUID) (GetMessageRow, error)
	GetLastMessageByChat(ctx context.Context, chatExternalID uuid.UUID) (GetLastMessageByChatRow, error)
	ListMessagesByChat(ctx context.Context, arg ListMessagesByChatParams) ([]ListMessagesByChatRow, error)
	ListMessagesByChatSince(ctx context.Context, arg ListMessagesByChatSinceParams) ([]ListMessagesByChatSinceRow, error)
	ListRecentMessagesByChat(ctx context.Context, arg ListRecentMessagesByChatParams) ([]ListRecentMessagesByChatRow, error)
	MarkMessageAsAdmin(ctx context.Context, messageExternalID uuid.UUID) (MarkMessageAsAdminRow, error)
	MarkMessageAsSystem(ctx context.Context, messageExternalID uuid.UUID) (MarkMessageAsSystemRow, error)
	AddOrUpdateReaction(ctx context.Context, arg AddOrUpdateReactionParams) (MessageReaction, error)
	InsertReactionWithWeight(ctx context.Context, arg InsertReactionWithWeightParams) (MessageReaction, error)
	RemoveReaction(ctx context.Context, arg RemoveReactionParams) error
	ToggleReaction(ctx context.Context, arg ToggleReactionParams) (ToggleReactionRow, error)
	CountAllReactionsByMessage(ctx context.Context, messageExternalID uuid.UUID) (int64, error)
	CountReactionsByMessage(ctx context.Context, messageExternalID uuid.UUID) ([]CountReactionsByMessageRow, error)
	CountUserReactions(ctx context.Context, userExternalID uuid.UUID) (int64, error)
	GetReactionsSummaryForChat(ctx context.Context, arg GetReactionsSummaryForChatParams) ([]GetReactionsSummaryForChatRow, error)
	GetTopReactionersInChat(ctx context.Context, arg GetTopReactionersInChatParams) ([]GetTopReactionersInChatRow, error)
	GetUserReactionForMessage(ctx context.Context, arg GetUserReactionForMessageParams) (MessageReaction, error)

	// Attachments
	CreateAttachment(ctx context.Context, arg CreateAttachmentParams) (MessageAttachment, error)
	DeleteAttachment(ctx context.Context, attachmentExternalID uuid.UUID) error
	DeleteAttachmentsByMessage(ctx context.Context, messageExternalID uuid.UUID) error
	GetAttachment(ctx context.Context, attachmentExternalID uuid.UUID) (MessageAttachment, error)
	GetAttachmentsMetadataByChat(ctx context.Context, arg GetAttachmentsMetadataByChatParams) ([]GetAttachmentsMetadataByChatRow, error)
	ListAllAttachmentsByMessage(ctx context.Context, messageExternalID uuid.UUID) ([]MessageAttachment, error)
	ListAttachmentsByChat(ctx context.Context, arg ListAttachmentsByChatParams) ([]MessageAttachment, error)
	ListAttachmentsByMessage(ctx context.Context, arg ListAttachmentsByMessageParams) ([]MessageAttachment, error)

	// Knowledge
	CreateKnowledge(ctx context.Context, arg CreateKnowledgeParams) (AiKnowledge, error)
	DeleteKnowledge(ctx context.Context, knowledgeExternalID uuid.UUID) error
	GetKnowledgeByID(ctx context.Context, knowledgeExternalID uuid.UUID) (AiKnowledge, error)
	SearchKnowledgeFulltext(ctx context.Context, arg SearchKnowledgeFulltextParams) ([]SearchKnowledgeFulltextRow, error)
	UpdateKnowledgeEmbedding(ctx context.Context, arg UpdateKnowledgeEmbeddingParams) error

	// Chunk
	CreateChunk(ctx context.Context, arg CreateChunkParams) (Chunk, error)
	DeleteChunk(ctx context.Context, chunkExternalID uuid.UUID) error
	GetChunkByID(ctx context.Context, chunkExternalID uuid.UUID) (Chunk, error)
	ListChunksBySource(ctx context.Context, sourceID pgtype.UUID) ([]Chunk, error)
	SearchChunksFulltext(ctx context.Context, arg SearchChunksFulltextParams) ([]SearchChunksFulltextRow, error)
	UpdateChunkEmbedding(ctx context.Context, arg UpdateChunkEmbeddingParams) error
	UpdateChunkStatus(ctx context.Context, arg UpdateChunkStatusParams) error

	// Session
	BlockSession(ctx context.Context, sessionExternalID uuid.UUID) error
	CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
	DeleteExpiredSessions(ctx context.Context) error
	DeleteSession(ctx context.Context, sessionExternalID uuid.UUID) error
	GetSessionByExternalID(ctx context.Context, sessionExternalID uuid.UUID) (Session, error)
	IsUserOnline(ctx context.Context, userExternalID uuid.UUID) (bool, error)
	ListSessionsByUser(ctx context.Context, userExternalID uuid.UUID) ([]Session, error)
	UpdateSessionToken(ctx context.Context, arg UpdateSessionTokenParams) (Session, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (Session, error)

	// SourceFile
	CreateSourceFile(ctx context.Context, arg CreateSourceFileParams) (SourceFile, error)
	GetSourceByExternalID(ctx context.Context, sourceExternalID uuid.UUID) (SourceFile, error)
	ListUploadedSources(ctx context.Context, limit int32) ([]SourceFile, error)
	MarkSourceProcessed(ctx context.Context, sourceID pgtype.Int8) error
}
