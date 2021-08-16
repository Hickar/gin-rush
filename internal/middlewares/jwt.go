package middlewares

import (
	"net/http"

	"github.com/Hickar/gin-rush/internal/security"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := http.StatusOK
		token := c.GetHeader("AUTHORIZATION")

		if token == "" {
			code = http.StatusUnauthorized
		} else {
			_, err := security.ParseJWT(security.TrimJWTPrefix(token))

			if err != nil {
				switch err.(*jwt.ValidationError).Errors {
				case jwt.ValidationErrorExpired:
					code = http.StatusUnauthorized
				default:
					code = http.StatusUnauthorized
				}
			}
		}

		if code != http.StatusOK {
			c.AbortWithStatus(code)
			return
		}

		c.Next()
	}
}
