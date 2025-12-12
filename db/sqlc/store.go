package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOpenChatAlreadyExists = errors.New("open chat already exists")

type CreateUserTxParams struct {
	CreateUserParams
}

type CreateUserTxResult struct {
	User CreateUserRow
}

type StartChatTxParams struct {
	UserExternalID uuid.UUID
	Content        string
}

type StartChatTxResult struct {
	Chat    Chat
	Message CreateMessageRow
}

type Store interface {
	Querier
	CreateChatTx(ctx context.Context, arg StartChatTxParams) (StartChatTxResult, error)
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	ToggleReactionTx(ctx context.Context, arg ToggleReactionParams) (ToggleReactionRow, error)
	InsertReactionTx(ctx context.Context, arg InsertReactionWithWeightParams) (MessageReaction, error)
}

type SQLStore struct {
	conn *pgxpool.Pool
	Querier
	*Queries
}

func NewStore(conn *pgxpool.Pool) *SQLStore {
	queries := New(conn)
	return &SQLStore{
		Querier: queries,
		conn:    conn,
		Queries: New(conn),
	}
}


func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("tx begin failed: %w", err)
	}

	q := store.WithTx(tx)

	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit failed: %w", err)
	}

	return nil
}

func (store *SQLStore) CreateChatTx(ctx context.Context, arg StartChatTxParams) (StartChatTxResult, error) {
	var result StartChatTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		chat, err := q.GetOpenChatByUser(ctx, arg.UserExternalID)
		if err == nil {
			result.Chat = chat
			return ErrOpenChatAlreadyExists
		} else if err != sql.ErrNoRows {
			return err
		}

		createChatArgs := CreateChatDefaultsParams{
			UserExternalID: arg.UserExternalID,
			Label:          "New Chat",
		}

		result.Chat, err = q.CreateChatDefaults(ctx, createChatArgs)
		if err != nil {
			return fmt.Errorf("failed to create chat: %w", err)
		}

		msgArg := CreateMessageParams{
			ChatExternalID:   result.Chat.ChatExternalID,
			SenderExternalID: arg.UserExternalID,
			Content:          arg.Content,
			IsSystemMessage:  false,
			IsAdminMessage:   false,
		}

		result.Message, err = q.CreateMessage(ctx, msgArg)
		if err != nil {
			return fmt.Errorf("failed to create initial message: %w", err)
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrOpenChatAlreadyExists) {
			return result, ErrOpenChatAlreadyExists
		}
		return StartChatTxResult{}, err
	}

	return result, nil
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}
		return nil
	})

	return result, err
}

func (store *SQLStore) ToggleReactionTx(ctx context.Context, arg ToggleReactionParams) (ToggleReactionRow, error) {
	var result ToggleReactionRow

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result, err = q.ToggleReaction(ctx, arg)
		if err != nil {
			return fmt.Errorf("toggle reaction failed: %w", err)
		}
		return nil
	})

	return result, err
}

func (store *SQLStore) InsertReactionTx(ctx context.Context, arg InsertReactionWithWeightParams) (MessageReaction, error) {
	var result MessageReaction

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result, err = q.InsertReactionWithWeight(ctx, arg)
		if err != nil {
			return fmt.Errorf("insert reaction failed: %w", err)
		}
		return nil
	})

	return result, err
}
