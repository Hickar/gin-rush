package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/gin-gonic/gin"
)

func TestJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conf := config.NewConfig("../../conf/config.test.json")
	r := gin.New()
	r.Use(JWT())
	r.GET("/endpoint", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		expectedCode := http.StatusUnauthorized

		req, _ := http.NewRequest("GET", "/endpoint", nil)
		req.Header.Set(conf.Server.JWTHeader, "Bearer invalidToken")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != expectedCode {
			t.Errorf("expected code %d, got %d instead", expectedCode, w.Code)
		}
	})

	t.Run("ValidToken", func(t *testing.T) {
		expectedCode := http.StatusOK
		token, _ := security.GenerateJWT(uint(0), conf.Server.JWTSecret)

		req, _ := http.NewRequest("GET", "/endpoint", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != expectedCode {
			t.Errorf("expected code %d, got %d instead", expectedCode, w.Code)
		}
	})
}