package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zahra-pzk/Chatbot_Project3/token"
)

func OptionalAuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.Next()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.Next()
			return
		}

		if strings.ToLower(fields[0]) != authorizationTypeBearer {
			ctx.Next()
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.Next()
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}