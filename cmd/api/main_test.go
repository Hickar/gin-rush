package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/router"
	"github.com/Hickar/gin-rush/pkg/database"
)

func TestServer(t *testing.T) {
	conf := config.NewConfig("../../conf/config.test.json")
	router := router.NewRouter(&conf.Server)
	database.NewMockDB()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	if err != nil {
		t.Errorf("Can't create Request object: %s", err)
	}

	if w.Code != 200 {
		t.Errorf("Got response with status code %d", w.Code)
	}
}
