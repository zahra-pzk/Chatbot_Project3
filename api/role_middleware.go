package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
)

const userObjectKey = "current_user_object"

func roleAuthMiddleware(store *db.SQLStore, allowedRoles []db.RoleType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get(authorizationPayloadKey)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("authorization payload not found")))
			return
		}

		payload, ok := val.(*token.Payload)
		if !ok || payload == nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("invalid authorization payload")))
			return
		}

		user, err := store.GetUserByExternalID(ctx, payload.UserExternalID)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("user not found")))
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		userRole := db.RoleType(user.Role)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				ctx.Set(userObjectKey, user)
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(errors.New("access denied: insufficient permissions")))
	}
}
