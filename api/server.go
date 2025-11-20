package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
)

type Server struct {
    store 	db.SQLStore
    router *gin.Engine
}

func NewServer(store *db.SQLStore) *Server{
	server := &Server{store: *store}
	router := gin.Default()

	router.POST("/chats", server.createChat)
	router.GET("/chats/:chatExternalID", server.getChat)
	router.GET("/chats", server.getChatsByUser)
	router.GET("/chats:chats", server.listChats)
	router.DELETE("/chats/:chatExternalID", server.deleteChat)
	router.PATCH("/chats/:chatExternalID/status", server.updateChatStatus)
	router.PATCH("/chats/:chatExternalID", server.updateChat)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}