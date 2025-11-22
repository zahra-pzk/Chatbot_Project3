package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
)

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func userOrAdminMiddleware(store *db.SQLStore, allowedRoles []db.RoleType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get(authorizationPayloadKey)
		if !ok {
			ctx.AbortWithStatusJSON(401, errorResponse(fmt.Errorf("authorization payload not found")))
			return
		}
		payload, ok := val.(*token.Payload)
		if !ok || payload == nil {
			ctx.AbortWithStatusJSON(401, errorResponse(fmt.Errorf("invalid authorization payload")))
			return
		}

		targetUserID := ctx.Param("userExternalID")

		userUUID, err := uuid.Parse(targetUserID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}

		if userUUID != payload.UserExternalID {
			ctx.JSON(http.StatusForbidden, errorResponse(errors.New("access denied")))
			return
		}

		user, err := store.GetUserByExternalID(ctx, payload.UserExternalID)
		if err != nil {
			ctx.AbortWithStatusJSON(500, errorResponse(err))
			return
		}
		userRole := db.RoleType(user.Role)
		for _, r := range allowedRoles {
			if userRole == r {
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatusJSON(403, errorResponse(fmt.Errorf("access denied: insufficient permissions")))
	}
}
