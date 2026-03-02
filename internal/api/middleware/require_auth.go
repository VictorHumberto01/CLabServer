package middleware

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vitub/CLabServer/internal/dtos"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

func RequireAuth(c *gin.Context) {
	tokenString, err := c.Cookie("Authorization")

	if err != nil {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
			err = nil
		}
	}

	if err != nil || tokenString == "" {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "Unauthorized - No token found",
		})
		c.Abort()
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		c.Abort()
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var user models.User
		if err := initializers.DB.First(&user, claims["sub"]).Error; err != nil {
			c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
			})
			c.Abort()
			return
		}

		if iatFloat, ok := claims["iat"].(float64); ok {
			iat := int64(iatFloat)
			// Allow a small grace period for timezone/clock sync (e.g., 5 seconds)
			if !user.PasswordChangedAt.IsZero() && iat < (user.PasswordChangedAt.Unix()-5) {
				c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
					Success: false,
					Error:   "Sessão expirada. Senha foi alterada recentemente.",
				})
				c.Abort()
				return
			}
		}

		c.Set("user", user)
		c.Next()
	} else {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		c.Abort()
	}
}
