package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hickar/gin-rush/internal/routes"
)

func TestServer(t *testing.T) {
	router := routes.Setup()

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
