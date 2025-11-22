package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

type Server struct {
	config     util.Config
	store      *db.SQLStore
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store *db.SQLStore) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/")
	authRoutes.Use(authMiddleware(server.tokenMaker))

	adminRoles := []db.RoleType{db.RoleTypeSuperadmin, db.RoleTypeAdmin}
	adminRoutes := authRoutes.Group("/admin")
	adminRoutes.Use(roleAuthMiddleware(server.store, adminRoles))

	adminRoutes.GET("/users/:userExternalID", server.getUser)
	adminRoutes.GET("/users", server.listUsers)
	adminRoutes.PUT("/users/:userExternalID", server.updateUser)
	adminRoutes.DELETE("/users/:userExternalID", server.deleteUser)

	authRoutes.PATCH("/users/:userExternalID/password",
		userOrAdminMiddleware(server.store, adminRoles),
		server.updatePassword)

	authRoutes.POST("/chats", server.createChat)
	authRoutes.GET("/chats/:chatExternalID", server.getChat)
	authRoutes.GET("/chats/user",
		userOrAdminMiddleware(server.store, adminRoles),
		server.getChatsByUser)
	adminRoutes.GET("/chats", server.listChats)
	adminRoutes.DELETE("/chats/:chatExternalID", server.deleteChat)
	authRoutes.PATCH("/chats/:chatExternalID/status", server.updateChatStatus)
	adminRoutes.PATCH("/chats/:chatExternalID", server.updateChat)

	authRoutes.POST("/messages", server.createMessage)
	authRoutes.GET("/chats/:chatExternalID/messages",
		userOrAdminMiddleware(server.store, adminRoles),
		server.listMessagesByChat)
	authRoutes.GET("/chats/:chatExternalID/messages/recent", server.listRecentMessagesByChat)
	authRoutes.GET("/messages/:messageExternalID", server.getMessage)
	adminRoutes.GET("/messages/:messageExternalID", server.getMessage)
	adminRoutes.PATCH("/messages/:messageExternalID", server.updateMessage)
	adminRoutes.DELETE("/messages/:messageExternalID", server.deleteMessage)

	server.router = router

}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}