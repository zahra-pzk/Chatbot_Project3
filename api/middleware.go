package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

func roleMiddleware(allowedRoles ...db.RoleType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

		userRole := db.RoleType(payload.Role)

		for _, role := range allowedRoles {
			if userRole == role {
				ctx.Next()
				return
			}
		}

		err := errors.New("user does not have the required permission")
		ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(err))
	}
}

func userOrAdminMiddleware(store *db.SQLStore, allowedAdminRoles []db.RoleType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

		paramIDStr := ctx.Param("userExternalID")
		paramID, err := uuid.Parse(paramIDStr)
		if err == nil && payload.UserExternalID == paramID {
			ctx.Next()
			return
		}

		userRole := db.RoleType(payload.Role)
		for _, role := range allowedAdminRoles {
			if userRole == role {
				ctx.Next()
				return
			}
		}

		err = errors.New("access denied: resource owner or admin role required")
		ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(err))
	}
}
