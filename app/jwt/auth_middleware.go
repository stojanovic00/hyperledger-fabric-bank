package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ExtractAndValidateToken(ctx)
		if err != nil {
			ctx.JSON(401, gin.H{"error": "Unauthorized - Invalid token"})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			ctx.JSON(401, gin.H{"error": "Unauthorized - Invalid token claims"})
			ctx.Abort()
			return
		}

		// Set the claims for further usage
		ctx.Set("userId", claims["userId"].(string))

		// Continue to the next middleware or handler
		ctx.Next()
	}
}
