package db

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)


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

func NewStore(conn *pgxpool.Pool) Store {
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
		return StartChatTxResult{}, err
	}

	return result, nil
}