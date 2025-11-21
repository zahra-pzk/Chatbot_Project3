package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

type Server struct {
	config		util.Config
    store 		db.SQLStore
	tokenMaker	token.Maker
    router 		*gin.Engine
}

func NewServer(config util.Config, store *db.SQLStore) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config: config,
		store: *store,
		tokenMaker: tokenMaker,
	}
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

    router.GET("/users/:userExternalID", server.getUser)
    router.GET("/users", server.listUsers) 
    router.PUT("/users/:userExternalID", server.updateUser)
    router.PATCH("/users/:userExternalID/password", server.updatePassword)
    router.DELETE("/users/:userExternalID", server.deleteUser)

	router.POST("/chats", server.createChat)
	router.GET("/chats/:chatExternalID", server.getChat)
	router.GET("/chats/user", server.getChatsByUser)
	router.GET("/chats", server.listChats)
	router.DELETE("/chats/:chatExternalID", server.deleteChat)
	router.PATCH("/chats/:chatExternalID/status", server.updateChatStatus)
	router.PATCH("/chats/:chatExternalID", server.updateChat)

	router.POST("/messages", server.createMessage)
    router.GET("/chats/:chatExternalID/messages", server.listMessagesByChat) 
    router.GET("/chats/:chatExternalID/messages/recent", server.listRecentMessagesByChat)
    router.GET("/messages/:messageExternalID", server.getMessage)
    router.PATCH("/messages/:messageExternalID", server.updateMessage)
    router.DELETE("/messages/:messageExternalID", server.deleteMessage)

	server.router = router
	return server, nil
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}