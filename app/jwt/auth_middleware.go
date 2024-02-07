package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthenticationMiddleware() gin.HandlerFunc {
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
		ctx.Set("role", claims["role"].(string))

		// Continue to the next middleware or handler
		ctx.Next()
	}
}

func AuthorizationMiddleware(requiredRole string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		providedRoleEntry, ok := ctx.Get("role")
		if !ok {
			ctx.JSON(400, gin.H{"error": "no auth parameters provided"})
			return
		}
		providedRole := providedRoleEntry.(string)

		if providedRole != requiredRole {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "you are not authorized for this action"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
