package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAuth returns a Gin middleware that enforces:
// 1. Valid Google-issued token
// 2. Email must be in allowedEmails (optional, pass nil to skip check)
func RequireAuth(allowedEmails []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetUserFromRequest(c.Request)
		if err != nil {
			fmt.Println("error: ", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if len(allowedEmails) > 0 && !contains(allowedEmails, user.Email) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// Store user info in context for downstream use
		c.Set("user", user)
		c.Next()
	}
}

func contains(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
