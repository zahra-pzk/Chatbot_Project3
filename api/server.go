package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/google/uuid"
	"github.com/zahra-pzk/Chatbot_Project3/api/ws"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

var adminChannelID = uuid.Nil

type Server struct {
	config     util.Config
	store      *db.SQLStore
	tokenMaker token.Maker
	router     *gin.Engine
	hub        *ws.Hub
}

func NewServer(config util.Config, store *db.SQLStore) (*Server, error) {
	hub := ws.NewHub()
	go hub.Run()
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		hub:        hub,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))
	router.SetTrustedProxies(nil)

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.GET("/ws/chats/:chatExternalID", server.ServeWs)
	router.GET("/ws/admin/chats", server.ServeAdminChatsWs)

	authRoutes := router.Group("/")
	authRoutes.Use(authMiddleware(server.tokenMaker))

	adminRoles := []db.RoleType{db.RoleTypeSuperadmin, db.RoleTypeAdmin}

	adminRoutes := authRoutes.Group("/admin")
	adminRoutes.Use(roleAuthMiddleware(server.store, adminRoles))

	authRoutes.GET("/users/:userExternalID", userOrAdminMiddleware(server.store, adminRoles), server.getUser)
	authRoutes.PATCH("/users/:userExternalID/password", userOrAdminMiddleware(server.store, adminRoles), server.updatePassword)

	adminRoutes.GET("/users", server.listUsers)
	adminRoutes.PUT("/users/:userExternalID", server.updateUser)
	adminRoutes.DELETE("/users/:userExternalID", server.deleteUser)

	authRoutes.POST("/chats", server.createChat)
	authRoutes.GET("/chats/:chatExternalID", server.getChat)
	authRoutes.GET("/chats/user", server.getChatsByUser)
	authRoutes.PATCH("/chats/:chatExternalID/status", server.updateChatStatus)

	adminRoutes.GET("/chats", server.listChats)
	adminRoutes.DELETE("/chats/:chatExternalID", server.deleteChat)
	adminRoutes.PATCH("/chats/:chatExternalID", server.updateChat)

	authRoutes.POST("/messages", server.createMessage)
	authRoutes.GET("/chats/:chatExternalID/messages", server.listMessagesByChat)
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
