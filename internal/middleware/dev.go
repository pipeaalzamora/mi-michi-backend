package middleware

import "github.com/gin-gonic/gin"

// DevAuth inyecta un userID fijo para desarrollo sin autenticación.
// NUNCA usar en producción.
func DevAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(UserIDKey, "dev-user-001")
		c.Next()
	}
}
