package middleware

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

func OptionalAuth(c *gin.Context) {
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
			err = nil
		}
	}

	if err != nil || tokenString == "" {
		tokenQueried := c.Query("token")
		if tokenQueried != "" {
			tokenString = tokenQueried
			err = nil
		}
	}

	if err != nil || tokenString == "" {
		fmt.Println("OptionalAuth: No token found")
		c.Next()
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		fmt.Printf("OptionalAuth: Token error: %v\n", err)
		c.Next()
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var user models.User
		if err := initializers.DB.First(&user, claims["sub"]).Error; err == nil {
			c.Set("user", user)
		}
	}

	c.Next()
}
