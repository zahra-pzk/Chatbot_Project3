package util

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func ErrorResponse(err error) gin.H {
	if err == nil {
		return gin.H{"error": "unknown error"}
	}
	timestamp := time.Now().Format(time.RFC3339)
	log.Printf("[%s] server error: %+v\n", timestamp, err)

	cause := errors.Cause(err)
	log.Printf("[%s] root cause: %v\n", timestamp, cause)

	return gin.H{
		"error": err.Error(),
		"time":  timestamp,
	}
}
