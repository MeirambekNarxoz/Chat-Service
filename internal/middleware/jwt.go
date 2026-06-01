package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func UserIDFromToken(tokenString, secret string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	if idFloat, ok := claims["user_id"].(float64); ok {
		return uint(idFloat), nil
	}

	// fallback: some tokens use "sub" as numeric user id string
	if sub, ok := claims["sub"].(string); ok {
		if id, err := strconv.ParseUint(sub, 10, 32); err == nil {
			return uint(id), nil
		}
	}
	if sub, ok := claims["sub"].(float64); ok {
		return uint(sub), nil
	}

	return 0, errors.New("user_id not found in token")
}

// GetUserID retrieves the UserID safely from the context
func GetUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := val.(uint)
	return userID, ok
}

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID uint
		var err error

		if userIDStr := c.GetHeader("X-User-Id"); userIDStr != "" {
			var id uint64
			id, err = strconv.ParseUint(userIDStr, 10, 32)
			if err == nil {
				userID = uint(id)
			}
		} else if secret != "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				userID, err = UserIDFromToken(strings.TrimSpace(strings.TrimPrefix(auth, "Bearer ")), secret)
			} else {
				err = errors.New("authorization header missing")
			}
		} else {
			err = errors.New("x-user-id header is missing")
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
