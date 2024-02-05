// utils.go in the utils package
package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"time"
)

var jwtSecret = []byte("your-secret-key")

func GenerateJWT(userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}

func ExtractAndValidateToken(ctx *gin.Context) (*jwt.Token, error) {
	tokenString := ctx.GetHeader("Authorization")
	if tokenString == "" {
		ctx.JSON(401, gin.H{"error": "Unauthorized - Token missing"})
		return nil, jwt.ErrSignatureInvalid
	}

	token, err := ValidateJWT(tokenString)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "Unauthorized - Invalid token"})
		return nil, jwt.ErrSignatureInvalid
	}

	return token, nil
}
