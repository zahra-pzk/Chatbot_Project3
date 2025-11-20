package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/zahra-pzk/Chatbot_Project3/api"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
)

const (
    dbSource      = "postgresql://zahra-pzk:25111380@localhost:5432/chatbot?sslmode=disable"
    serverAddress = "0.0.0.0:8080"
)

func main() {
    ctx := context.Background()

    pool, err := pgxpool.New(ctx, dbSource)
    if err != nil {
        log.Fatal("cannot connect to db:", err)
    }

    store := db.NewStore(pool)
    server := api.NewServer(store)

    if err := server.Start(serverAddress); err != nil {
        log.Fatal("cannot start server:", err)
    }
}
