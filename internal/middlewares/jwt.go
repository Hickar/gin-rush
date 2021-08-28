package middlewares

import (
	"net/http"

	"github.com/Hickar/gin-rush/internal/security"
	"github.com/gin-gonic/gin"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var claims *security.Claims
		var err error

		code := http.StatusOK
		token := trimJWTPrefix(c.GetHeader("AUTHORIZATION"))

		if token == "" {
			code = http.StatusUnauthorized
		} else {
			claims, err = security.ParseJWT(token)

			if err != nil {
				code = http.StatusUnauthorized
			}
		}

		if code != http.StatusOK {
			c.AbortWithStatus(code)
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func trimJWTPrefix(header string) string {
	return header[7:]
}
