package route

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
	"github.com/zahra-pzk/Chatbot_Project3/util"
	"github.com/zahra-pzk/Chatbot_Project3/api/handler"
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
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))
	router.SetTrustedProxies(nil)
	authHandler := handler.NewAuthHandler(server.store, server.tokenMaker, server.config)
	router.POST("/users", authHandler.CreateUser)
	router.POST("/users/guest", authHandler.CreateGuest)
	router.POST("/users/login/email", authHandler.LoginUserByEmail)
	router.POST("/users/login/username", authHandler.LoginUserByUsernsme)
	router.POST("/users/login/phone", authHandler.LoginUserByPhone)
	router.POST("/tokens/renew_access", authHandler.RenewAccessToken)
	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
