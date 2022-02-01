package middleware

import (
	"net/http"
	"strings"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/gin-gonic/gin"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var claims *security.Claims
		var err error

		conf := config.GetConfig().Server
		code := http.StatusOK
		token := trimJWTPrefix(conf.JWTBearerPrefix, c.GetHeader(conf.JWTHeader))

		if token == "" {
			code = http.StatusUnauthorized
		} else {
			claims, err = security.ParseJWT(token, conf.JWTSecret)

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

func trimJWTPrefix(prefix, header string) string {
	return strings.Trim(header[len(prefix):], " ")
}
