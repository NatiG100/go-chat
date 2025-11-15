package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"example.com/go-chat/internal/drivers"
)

func AuthMiddleware(jwt *drivers.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}

		// remove Bearer prefix
		if after, ok :=strings.CutPrefix(auth, "Bearer "); ok  {
			auth = after
		}

		// verify token → returns uuid.UUID
		userID, err := jwt.Verify(auth)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// store uuid directly (NO type assertion)
		if userID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
