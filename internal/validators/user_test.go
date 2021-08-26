package validators

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hickar/gin-rush/internal/api"
	"github.com/Hickar/gin-rush/pkg/request"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func TestUserCredentialsValidators(t *testing.T) {
	r := gin.New()
	r.POST("/api/user", api.CreateUser)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notblank", NotBlank)
		v.RegisterValidation("validemail", ValidEmail)
		v.RegisterValidation("validpassword", ValidPassword)
	}

	tests := []struct {
		Name         string
		BodyData     request.CreateUserRequest
		ExpectedCode int
		Msg          string
	}{
		{
			"Blank name",
			request.CreateUserRequest{Name: "", Email: "dummy@email.io", Password: "Pass/w0rd"},
			http.StatusUnprocessableEntity,
			"Should return 422, name is blank",
		},
		{
			"Blank email",
			request.CreateUserRequest{Name: "someUser", Email: "", Password: "Pass/w0rd"},
			http.StatusUnprocessableEntity,
			"Should return 422, email is blank",
		},
		{
			"Invalid email",
			request.CreateUserRequest{Name: "someUser", Email: "invalid.email", Password: "Pass/w0rd"},
			http.StatusUnprocessableEntity,
			"Should return 422, specified email doesn't meet requirements of RFC 3696 standard",
		},
		{
			"Invalid password",
			request.CreateUserRequest{Name: "someUser", Email: "dummy@email.io", Password: "v"},
			http.StatusUnprocessableEntity,
			"Must return 422, password must be length of 8, contain one uppercase character, symbol and digit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			bodyData, _ := json.Marshal(tt.BodyData)
			buf := bytes.NewBuffer(bodyData)

			req, err := http.NewRequest("POST", "/api/user", buf)
			req.Header.Set("Content-Type", "application/json")

			if err != nil {
				t.Fatalf("Error during request construction: %s", err)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.ExpectedCode {
				t.Errorf("Returned code %d – %s", w.Code, tt.Msg)
			}
		})
	}
}