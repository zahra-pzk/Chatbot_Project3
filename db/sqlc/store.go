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
type StartChatTxParams struct {
	UserExternalID uuid.UUID
	Content        string
}

type StartChatTxResult struct {
	Chat    Chat
	Message Message
}

type Store interface {
	Querier
    CreateChatTx(ctx context.Context, arg StartChatTxParams) (StartChatTxResult, error) 
}

type SQLStore struct {
	conn *pgxpool.Pool
	*Queries
}

func NewStore(conn *pgxpool.Pool) *SQLStore {
	return &SQLStore{
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

	result.Chat, err = q.CreateChatDefaults(ctx, arg.UserExternalID)
	if err != nil {
		return fmt.Errorf("store: failed to create chat: %w", err)
	}

	msgArg := CreateMessageParams{
		ChatExternalID:   result.Chat.ChatExternalID,
		SenderExternalID: arg.UserExternalID,
		Content:          arg.Content,
		IsSystemMessage:  false,
	}

	result.Message, err = q.CreateMessage(ctx, msgArg)
	if err != nil {
		return fmt.Errorf("store: failed to create initial message: %w", err)
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