package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 从环境变量获取JWT密钥
		jwtSecret := os.Getenv("WALLET_JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "your-secret-key-here-change-in-production"
		}

		// 解析JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 从token中获取用户信息
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			rawUserID, exists := claims["user_id"]
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}

			// 规范化 user_id 为 uint64
			var userIDUint uint64
			switch v := rawUserID.(type) {
			case float64:
				userIDUint = uint64(v)
			case int:
				userIDUint = uint64(v)
			case int32:
				userIDUint = uint64(v)
			case int64:
				userIDUint = uint64(v)
			case uint:
				userIDUint = uint64(v)
			case uint32:
				userIDUint = uint64(v)
			case uint64:
				userIDUint = v
			case string:
				if parsed, perr := strconv.ParseUint(v, 10, 64); perr == nil {
					userIDUint = parsed
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token user_id"})
					c.Abort()
					return
				}
			default:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token user_id type"})
				c.Abort()
				return
			}

			c.Set("user_id", userIDUint)
		}

		c.Next()
	}
}
